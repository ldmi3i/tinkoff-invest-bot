package avr

import (
	"encoding/json"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type algoState struct {
	InitAmount map[string]int64           `json:"initAmount"` //Initial lot amount of instrument
	BuyPrice   map[string]decimal.Decimal `json:"buyPrice"`   //Buy price - sell no cheaper then it
}

func configure(confCtx map[string]string, state *algoState, logger *zap.SugaredLogger) error {
	if state.InitAmount == nil {
		state.InitAmount = make(map[string]int64)
	}
	if state.BuyPrice == nil {
		state.BuyPrice = make(map[string]decimal.Decimal)
	}
	instrData, ok := confCtx[dto.InstrAmountField]
	if ok {
		var instrAmount dto.InstrumentsInfo
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
