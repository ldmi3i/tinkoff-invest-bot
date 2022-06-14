package dto

import "github.com/shopspring/decimal"

const (
	InstrAmountField string = "instrAmount"
)

type InstrumentsInfo struct {
	Instruments []*InstrumentInfo `json:"instruments"`
}

type InstrumentInfo struct {
	Figi        string          `json:"figi"`     //Instrument figi
	Amount      int64           `json:"amount"`   //Amount of instrument available
	BuyPosPrice decimal.Decimal `json:"buyPrice"` //To specify instrument bought price if it is
}
