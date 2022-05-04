package avr

import (
	"github.com/shopspring/decimal"
	"time"
)

type DataProc interface {
	GetDataStream() (<-chan procData, error)
	Go()
	Stop() error
}

type procData struct {
	Figi string
	Time time.Time
	LAV  decimal.Decimal //average by long window
	SAV  decimal.Decimal //average by short window
}

const (
	ShortDur string = "short_dur"
	LongDur  string = "long_dur"
)
