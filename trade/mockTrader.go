package trade

import (
	"github.com/shopspring/decimal"
	"invest-robot/domain"
	"invest-robot/dto"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/strategy/stmodel"
	"invest-robot/trade/trmodel"
	"log"
	"time"
)

type MockTrader struct {
	hRep         repository.HistoryRepository
	sub          *stmodel.Subscription
	statCh       chan dto.HistStatResponse
	logs         map[string]int64        //number of positions per buy
	figiCurrency map[string]string       //currency to instrument figi relation
	figiHist     map[string][]histRecord //history of each figi - to convenience interpolation
}

type histRecord struct {
	Time  time.Time
	Figi  string
	Price decimal.Decimal
}

//auxiliary to keep trader data information
type mockTraderData struct {
	LastTime  time.Time
	ResAmount map[string]decimal.Decimal
	ResInstr  map[string]int64
	BuyOper   uint
	SellOper  uint
}

func (t *MockTrader) Go() {
	go t.procBg()
}

func (t *MockTrader) AddSubscription(sub *stmodel.Subscription) error {
	if err := t.populateHistory(); err != nil {
		return err
	}
	if t.sub != nil {
		return errors.NewDoubleSubErr("Double subscription not available for mock trader")
	}
	t.sub = sub
	return nil
}

func (t *MockTrader) procBg() {
	trDat := mockTraderData{
		LastTime:  time.Time{},
		ResAmount: make(map[string]decimal.Decimal),
		ResInstr:  make(map[string]int64),
		BuyOper:   0,
		SellOper:  0,
	}

	for act := range t.sub.AChan {
		action := act.Action
		trDat.LastTime = action.RetrievedAt
		actCurrency, exst := t.figiCurrency[action.InstrFigi]
		if !exst {
			log.Printf("Requested unexpected figi: %s", action.InstrFigi)
			t.sub.RChan <- t.getRespWithStatus(action, domain.FAILED)
			continue
		}
		opInfo := trmodel.OpInfo{Currency: actCurrency}
		action.Currency = actCurrency
		opInfo.Lim = act.GetCurrLimit(actCurrency)
		opInfo.PosNum = t.logs[action.InstrFigi]
		var err error
		opInfo.LotPrice, err = t.calcPrice(action.InstrFigi, action.RetrievedAt)
		if err != nil {
			log.Printf("Error while calculating figi price: %s", err)
			t.sub.RChan <- t.getRespWithStatus(action, domain.FAILED)
			continue
		}
		if opInfo.Lim.IsZero() || opInfo.PosNum == 0 || opInfo.LotPrice.IsZero() {
			log.Printf("Limit or lot price is zero; figi: %s; limit: %s; lot num: %d;lot price: %s",
				action.InstrFigi, opInfo.Lim, opInfo.PosNum, opInfo.LotPrice)
			t.sub.RChan <- t.getRespWithStatus(action, domain.FAILED)
			continue
		}
		if action.Direction == domain.BUY {
			t.procBuy(opInfo, action, &trDat)
		} else {
			t.procSell(opInfo, action, &trDat)
		}
	}

	log.Println("Action channel closed, stopping mock trader...")
	stat := t.calcMoneyStat(&trDat)
	stat.SellOpNum = trDat.SellOper
	stat.BuyOpNum = trDat.BuyOper
	t.statCh <- stat
}

func (t *MockTrader) procBuy(opInfo trmodel.OpInfo, action *domain.Action, trDat *mockTraderData) {
	onePrice := decimal.NewFromInt(opInfo.PosNum).Mul(opInfo.LotPrice)
	if onePrice.GreaterThan(opInfo.Lim) {
		log.Printf("Not enough money for figi %s; limit: %s; lot price: %s; one price: %s",
			action.InstrFigi, opInfo.Lim, opInfo.LotPrice, onePrice)
		t.sub.RChan <- t.getRespWithStatus(action, domain.FAILED)
	}
	operNum := opInfo.Lim.Div(onePrice).Floor()
	moneyAmount := operNum.Mul(onePrice)
	instrAmount := operNum.IntPart() * opInfo.PosNum
	trDat.ResInstr[action.InstrFigi] = trDat.ResInstr[action.InstrFigi] + instrAmount
	trDat.ResAmount[opInfo.Currency] = trDat.ResAmount[opInfo.Currency].Sub(moneyAmount)
	action.Amount = moneyAmount
	action.InstrAmount = instrAmount
	trDat.BuyOper += 1
	t.sub.RChan <- t.getRespWithStatus(action, domain.SUCCESS)
}

func (t *MockTrader) procSell(opInfo trmodel.OpInfo, action *domain.Action, trDat *mockTraderData) {
	if action.InstrAmount == 0 {
		log.Println("InstrAmount is 0 - nothing to sell")
		t.sub.RChan <- t.getRespWithStatus(action, domain.FAILED)
	}
	if action.InstrAmount < opInfo.PosNum {
		log.Printf("Not enough lots for one operation; requested: %d; lot num: %d", action.InstrAmount, opInfo.PosNum)
		t.sub.RChan <- t.getRespWithStatus(action, domain.FAILED)
	}
	price, err := t.calcPrice(action.InstrFigi, action.RetrievedAt)
	if err != nil {
		log.Println("Can't resolve price by figi, canceling operation...")
		t.sub.RChan <- t.getRespWithStatus(action, domain.FAILED)
	}
	fullPart := (action.InstrAmount / opInfo.PosNum) * opInfo.PosNum
	moneyAmount := price.Mul(decimal.NewFromInt(fullPart))
	trDat.ResAmount[opInfo.Currency] = trDat.ResAmount[opInfo.Currency].Add(moneyAmount)
	trDat.ResInstr[action.InstrFigi] = trDat.ResInstr[action.InstrFigi] - fullPart
	action.Amount = moneyAmount
	trDat.SellOper += 1
	t.sub.RChan <- t.getRespWithStatus(action, domain.SUCCESS)
}

func (t MockTrader) getRespWithStatus(action *domain.Action, status domain.ActionStatus) *stmodel.ActionResp {
	action.Status = status
	return &stmodel.ActionResp{Action: action}
}

func (t MockTrader) calcMoneyStat(trDat *mockTraderData) dto.HistStatResponse {
	//calculate money values according to current rate
	for figi, amount := range trDat.ResInstr {
		if amount != 0 {
			lotPrice, err := t.calcPrice(figi, trDat.LastTime)
			if err != nil {
				log.Printf("Error whle calculating price; figi: %s; time: %s", figi, trDat.LastTime)
			}
			currency, exst := t.figiCurrency[figi]
			if exst && err == nil {
				trDat.ResAmount[currency] = trDat.ResAmount[currency].Add(lotPrice.Mul(decimal.NewFromInt(amount)))
			}
		}
	}
	return dto.HistStatResponse{CurBalance: trDat.ResAmount}
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

func NewMockTrader(hRep repository.HistoryRepository, lots map[string]int64, figiCurrency map[string]string) MockTrader {
	log.Printf("Initializing mock trader with currencies: %+v , lot nums: %+v", lots, figiCurrency)
	return MockTrader{statCh: make(chan dto.HistStatResponse), hRep: hRep, logs: lots, figiCurrency: figiCurrency}
}
