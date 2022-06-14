package trade

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/errors"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/repository"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/strategy/stmodel"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/trade/trmodel"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"time"
)

type MockTrader struct {
	hRep         repository.HistoryRepository
	sub          *stmodel.Subscription
	statCh       chan dto.HistStatResponse
	lots         map[string]int64        //number of positions per buy
	figiCurrency map[string]string       //currency to instrument figi relation
	figiHist     map[string][]histRecord //history of each figi - to convenience interpolation
	logger       *zap.SugaredLogger
	ctx          context.Context
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

func (t *MockTrader) Go(ctx context.Context) {
	t.ctx = ctx
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

OUT:
	for {
		select {
		case <-t.ctx.Done():
			t.logger.Info("Mock trader cancel request received...")
			break OUT
		case act, ok := <-t.sub.AChan:
			if !ok {
				t.logger.Info("Incoming stream closed, stopping")
				break OUT
			}
			action := act.Action
			trDat.LastTime = action.RetrievedAt
			actCurrency, exst := t.figiCurrency[action.InstrFigi]
			if !exst {
				t.logger.Warnf("Requested unexpected figi: %s", action.InstrFigi)
				t.sub.RChan <- t.getRespWithStatus(action, entity.Failed)
				continue
			}
			opInfo := trmodel.OpInfo{Currency: actCurrency}
			action.Currency = actCurrency
			opInfo.Lim = act.GetCurrLimit(actCurrency)
			opInfo.PosInLot = t.lots[action.InstrFigi]
			var err error
			opInfo.PosPrice, err = t.calcPrice(action.InstrFigi, action.RetrievedAt)
			if err != nil {
				t.logger.Errorf("Error while calculating figi price: %s", err)
				t.sub.RChan <- t.getRespWithStatus(action, entity.Failed)
				continue
			}
			if opInfo.Lim.IsZero() || opInfo.PosInLot == 0 || opInfo.PosPrice.IsZero() {
				t.logger.Warnf("Limit or lot price is zero; figi: %s; limit: %s; pos in lot: %d;lot price: %s",
					action.InstrFigi, opInfo.Lim, opInfo.PosInLot, opInfo.PosPrice)
				t.sub.RChan <- t.getRespWithStatus(action, entity.Failed)
				continue
			}
			if action.Direction == entity.Buy {
				t.procBuy(opInfo, action, &trDat)
			} else {
				t.procSell(opInfo, action, &trDat)
			}
		}
	}

	t.logger.Info("Action channel closed, stopping mock trader...")
	stat := t.calcMoneyStat(&trDat)
	stat.SellOpNum = trDat.SellOper
	stat.BuyOpNum = trDat.BuyOper
	t.statCh <- stat
	close(t.sub.RChan)
}

func (t *MockTrader) procBuy(opInfo trmodel.OpInfo, action *entity.Action, trDat *mockTraderData) {
	lotPrice := decimal.NewFromInt(opInfo.PosInLot).Mul(opInfo.PosPrice)
	if lotPrice.GreaterThan(opInfo.Lim) {
		t.logger.Infof("Not enough money for figi %s; limit: %s; lot price: %s; one price: %s",
			action.InstrFigi, opInfo.Lim, opInfo.PosPrice, lotPrice)
		t.sub.RChan <- t.getRespWithStatus(action, entity.Failed)
		return
	}
	lotNum := opInfo.Lim.Div(lotPrice).Floor()
	moneyAmount := lotNum.Mul(lotPrice)
	instrAmount := lotNum.IntPart()
	trDat.ResInstr[action.InstrFigi] = trDat.ResInstr[action.InstrFigi] + instrAmount
	trDat.ResAmount[opInfo.Currency] = trDat.ResAmount[opInfo.Currency].Sub(moneyAmount)
	action.TotalPrice = moneyAmount
	action.LotAmount = instrAmount
	action.PositionPrice = opInfo.PosPrice
	action.LotsExecuted = instrAmount
	trDat.BuyOper += 1
	t.sub.RChan <- t.getRespWithStatus(action, entity.Success)
}

func (t *MockTrader) procSell(opInfo trmodel.OpInfo, action *entity.Action, trDat *mockTraderData) {
	if action.LotAmount == 0 {
		t.logger.Info("LotAmount is 0 - nothing to sell")
		t.sub.RChan <- t.getRespWithStatus(action, entity.Failed)
		return
	}
	price, err := t.calcPrice(action.InstrFigi, action.RetrievedAt)
	if err != nil {
		t.logger.Error("Can't resolve price by figi, canceling operation...")
		t.sub.RChan <- t.getRespWithStatus(action, entity.Failed)
		return
	}
	moneyAmount := price.Mul(decimal.NewFromInt(action.LotAmount * opInfo.PosInLot)) //Money amount is a price multiplied by num of positions
	trDat.ResAmount[opInfo.Currency] = trDat.ResAmount[opInfo.Currency].Add(moneyAmount)
	trDat.ResInstr[action.InstrFigi] = trDat.ResInstr[action.InstrFigi] - action.LotAmount
	//Negative amount of instrument not allowed, means initial amount of instrument existed
	if trDat.ResInstr[action.InstrFigi] < 0 {
		trDat.ResInstr[action.InstrFigi] = 0
	}
	action.TotalPrice = moneyAmount
	trDat.SellOper += 1
	t.sub.RChan <- t.getRespWithStatus(action, entity.Success)
}

func (t MockTrader) getRespWithStatus(action *entity.Action, status entity.ActionStatus) *stmodel.ActionResp {
	action.Status = status
	return &stmodel.ActionResp{Action: action}
}

func (t MockTrader) calcMoneyStat(trDat *mockTraderData) dto.HistStatResponse {
	//calculate money values according to current rate
	for figi, amount := range trDat.ResInstr {
		if amount != 0 {
			posPrice, err := t.calcPrice(figi, trDat.LastTime)
			if err != nil {
				t.logger.Errorf("Error whle calculating price; figi: %s; time: %s", figi, trDat.LastTime)
			}
			posInLot, ok := t.lots[figi]
			if !ok {
				posInLot = 1
			}
			lotPrice := posPrice.Mul(decimal.NewFromInt(posInLot))
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
	t.logger.Debugf("Interpolate between up(val,T): (%s,%s) down (%s,%s) with result (%s,%s)",
		uppVal, uppTm, lwrVal, lwrTm, res, tm)
	return res, nil
}

func (t *MockTrader) RemoveSubscription(id uint) error {
	return errors.NewNotImplemented()
}

func (t MockTrader) GetStatCh() chan dto.HistStatResponse {
	return t.statCh
}

func NewMockTrader(hRep repository.HistoryRepository, lots map[string]int64, figiCurrency map[string]string, logger *zap.SugaredLogger) MockTrader {
	logger.Debugf("Initializing mock trader with currencies: %+v , lot nums: %+v", lots, figiCurrency)
	return MockTrader{statCh: make(chan dto.HistStatResponse), hRep: hRep, lots: lots, figiCurrency: figiCurrency, logger: logger}
}
