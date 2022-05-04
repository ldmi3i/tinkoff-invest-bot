package trade

import (
	"github.com/shopspring/decimal"
	"invest-robot/domain"
	"invest-robot/dto"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/strategy/model"
	"log"
	"time"
)

type MockTrader struct {
	hRep         repository.HistoryRepository
	sub          *model.Subscription
	statCh       chan dto.HistStatResponse
	lotNums      map[string]int64        //number of lots per buy
	figiCurrency map[string]string       //currency to instrument figi relation
	figiHist     map[string][]histRecord //history of each figi - to convenience interpolation
}

type histRecord struct {
	Time  time.Time
	Figi  string
	Price decimal.Decimal
}

func (t *MockTrader) AddSubscription(sub *model.Subscription) error {
	if err := t.populateHistory(); err != nil {
		return err
	}
	if t.sub != nil {
		return errors.NewDoubleSubErr("Double subscription not available for mock trader")
	}
	t.sub = sub
	go t.procBg()
	return nil
}

func (t *MockTrader) procBg() {
	var buyOper uint = 0
	var sellOper uint = 0
	lastTime := time.Time{}
	resAmount := make(map[string]decimal.Decimal)
	resInstr := make(map[string]int64)

	for act := range t.sub.AChan {
		action := act.Action
		lastTime = action.RetrievedAt
		actCurrency, exst := t.figiCurrency[action.InstrFigi]
		action.Currency = actCurrency
		if !exst {
			log.Printf("Requested unexpected figi: %s", action.InstrFigi)
			action.Status = domain.FAILED
			resp := model.ActionResp{
				IsSuccess: false,
				Action:    action,
			}
			t.sub.RChan <- resp
			continue
		}
		lim := act.GetCurrLimit(actCurrency)
		lotNum := t.lotNums[action.InstrFigi]
		lotPrice, err := t.calcPrice(action.InstrFigi, action.RetrievedAt)
		if err != nil {
			log.Printf("Error while calculating figi price: %s", err)
			action.Status = domain.FAILED
			resp := model.ActionResp{
				IsSuccess: false,
				Action:    action,
			}
			t.sub.RChan <- resp
			continue
		}
		if lim.IsZero() || lotNum == 0 || lotPrice.IsZero() {
			log.Printf("Limit or lot price is zero; figi: %s; limit: %s; lot num: %d;lot price: %s",
				action.InstrFigi, lim, lotNum, lotPrice)
			action.Status = domain.FAILED
			resp := model.ActionResp{
				IsSuccess: false,
				Action:    action,
			}
			t.sub.RChan <- resp
			continue
		}
		if action.Direction == domain.BUY {
			onePrice := decimal.NewFromInt(lotNum).Mul(lotPrice)
			if onePrice.GreaterThan(lim) {
				log.Printf("Not enough money for figi %s; limit: %s; lot price: %s; one price: %s",
					action.InstrFigi, lim, lotPrice, onePrice)
				action.Status = domain.FAILED
				resp := model.ActionResp{
					IsSuccess: false,
					Action:    action,
				}
				t.sub.RChan <- resp
			}
			operNum := lim.Div(onePrice).Floor()
			moneyAmount := operNum.Mul(onePrice)
			instrAmount := operNum.IntPart() * lotNum
			resInstr[action.InstrFigi] = resInstr[action.InstrFigi] + instrAmount
			resAmount[actCurrency] = resAmount[actCurrency].Sub(moneyAmount)
			buyOper += 1
			action.Status = domain.SUCCESS
			action.Amount = moneyAmount
			action.InstrAmount = instrAmount
			t.sub.RChan <- model.ActionResp{
				IsSuccess: true,
				Action:    action,
			}
		} else {
			if action.InstrAmount == 0 {
				log.Println("InstrAmount is 0 - nothing to sell")
				action.Status = domain.FAILED
				resp := model.ActionResp{
					IsSuccess: false,
					Action:    action,
				}
				t.sub.RChan <- resp
			}
			if action.InstrAmount < lotNum {
				log.Printf("Not enough lots for one operation; requested: %d; lot num: %d", action.InstrAmount, lotNum)
				action.Status = domain.FAILED
				resp := model.ActionResp{
					IsSuccess: false,
					Action:    action,
				}
				t.sub.RChan <- resp
			}
			price, err := t.calcPrice(action.InstrFigi, action.RetrievedAt)
			if err != nil {
				log.Println("Can't resolve price by figi, canceling operation...")
				action.Status = domain.FAILED
				resp := model.ActionResp{
					IsSuccess: false,
					Action:    action,
				}
				t.sub.RChan <- resp
			}
			fullPart := (action.InstrAmount / lotNum) * lotNum
			moneyAmount := price.Mul(decimal.NewFromInt(fullPart))
			resAmount[actCurrency] = resAmount[actCurrency].Add(moneyAmount)
			resInstr[action.InstrFigi] = resInstr[action.InstrFigi] - fullPart
			sellOper += 1
			action.Status = domain.SUCCESS
			action.Currency = actCurrency
			action.Amount = moneyAmount
			t.sub.RChan <- model.ActionResp{
				IsSuccess: true,
				Action:    action,
			}
		}
	}

	log.Println("Action channel closed, stopping mock trader...")
	stat := t.calcMoneyStat(lastTime, resAmount, resInstr)
	stat.SellOpNum = sellOper
	stat.BuyOpNum = buyOper
	t.statCh <- stat
}

func (t MockTrader) calcMoneyStat(lastTime time.Time, resAmount map[string]decimal.Decimal, resInstr map[string]int64) dto.HistStatResponse {
	//calculate money values according to current rate
	for figi, amount := range resInstr {
		if amount != 0 {
			lotPrice, err := t.calcPrice(figi, lastTime)
			if err != nil {
				log.Printf("Error whle calculating price; figi: %s; time: %s", figi, lastTime)
			}
			currency, exst := t.figiCurrency[figi]
			if exst && err == nil {
				resAmount[currency] = resAmount[currency].Add(lotPrice.Mul(decimal.NewFromInt(amount)))
			}
		}
	}
	return dto.HistStatResponse{CurBalance: resAmount}
}

func (t *MockTrader) populateHistory() error {
	figis := make([]string, 0, len(t.figiCurrency))
	for figi := range t.figiCurrency {
		figis = append(figis, figi)
	}
	history, err := t.hRep.FindAllByFigis(figis)
	if err != nil {
		return err
	}
	t.figiHist = make(map[string][]histRecord)
	for _, figi := range figis {
		t.figiHist[figi] = make([]histRecord, 0)
	}
	for _, hRec := range history {
		rec := histRecord{
			Time:  hRec.Time,
			Figi:  hRec.Figi,
			Price: hRec.Close,
		}
		t.figiHist[hRec.Figi] = append(t.figiHist[hRec.Figi], rec)
	}
	return nil
}

func (t MockTrader) calcPrice(figi string, tm time.Time) (decimal.Decimal, error) {
	hist, exst := t.figiHist[figi]
	if !exst {
		return decimal.Zero, errors.NewUnexpectedError("Requested figi not found")
	}
	uppTm := time.Now()
	uppVal := decimal.Zero
	lwrTm := time.Time{}
	lwrVal := decimal.Zero
	//TODO make hist sorted and use binary search
	for _, hRec := range hist {
		if hRec.Time.Equal(tm) {
			return hRec.Price, nil
		} else if hRec.Time.After(tm) && hRec.Time.Before(uppTm) {
			uppTm = hRec.Time
			uppVal = hRec.Price
		} else if hRec.Time.Before(tm) && hRec.Time.After(lwrTm) {
			lwrTm = hRec.Time
			lwrVal = hRec.Price
		}
	}
	uppUnx := decimal.NewFromInt(uppTm.Unix())
	lwrUnx := decimal.NewFromInt(lwrTm.Unix())
	search := decimal.NewFromInt(tm.Unix())
	res := lwrVal.Add(uppVal.Sub(lwrVal).Mul(search.Sub(lwrUnx)).Div(uppUnx.Sub(lwrUnx)))
	log.Printf("Interpolate between up(val,T): (%s,%s) down (%s,%s) with result (%s,%s)",
		uppVal, uppTm, lwrVal, lwrTm, res, tm)
	return res, nil
}

func (t *MockTrader) RemoveSubscription(id uint) error {
	return errors.NewNotImplemented()
}

func (t MockTrader) GetStatCh() chan dto.HistStatResponse {
	return t.statCh
}

func NewMockTrader(hRep repository.HistoryRepository, lotNums map[string]int64, figiCurrency map[string]string) MockTrader {
	log.Printf("Initializing mock trader with currencies: %+v , lot nums: %+v", lotNums, figiCurrency)
	return MockTrader{statCh: make(chan dto.HistStatResponse), hRep: hRep, lotNums: lotNums, figiCurrency: figiCurrency}
}
