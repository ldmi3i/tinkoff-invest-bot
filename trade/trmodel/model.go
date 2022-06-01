package trmodel

import (
	"github.com/shopspring/decimal"
	"time"
)

type OpInfo struct {
	PosInLot  int64
	PosPrice  decimal.Decimal
	Lim       decimal.Decimal
	PriceStep decimal.Decimal
	Currency  string
}

type Timed[T any] struct {
	Data T
	Time time.Time
}
