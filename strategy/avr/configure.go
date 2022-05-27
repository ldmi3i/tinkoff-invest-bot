package avr

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type instrumentsInfo struct {
	Instruments []*instrumentInfo `json:"instruments"`
}

type instrumentInfo struct {
	Figi        string          `json:"figi"`     //Instrument figi
	Amount      int64           `json:"amount"`   //Amount of instrument available
	BuyPosPrice decimal.Decimal `json:"buyPrice"` //To specify instrument bought price if it is
}

type algoState struct {
	InitAmount map[string]int64           `json:"initAmount"` //Initial lot amount of instrument
	BuyPrice   map[string]decimal.Decimal `json:"buyPrice"`   //Buy price - sell no cheaper than it
}

const (
	instrAmountField string = "instrAmount"
)

func configure(confCtx map[string]string, state *algoState, logger *zap.SugaredLogger) error {
	if state.InitAmount == nil {
		state.InitAmount = make(map[string]int64)
	}
	if state.BuyPrice == nil {
		state.BuyPrice = make(map[string]decimal.Decimal)
	}
	instrData, ok := confCtx[instrAmountField]
	if ok {
		var instrAmount instrumentsInfo
		err := json.Unmarshal([]byte(instrData), &instrAmount)
		if err != nil {
			logger.Warnf("Unable unmarshal '%s' to Instruments", instrData)
			return err
		} else {
			for _, instr := range instrAmount.Instruments {
				state.InitAmount[instr.Figi] = instr.Amount
				if !instr.BuyPosPrice.IsZero() {
					state.BuyPrice[instr.Figi] = instr.BuyPosPrice
				}
			}
		}
	}
	return nil
}
