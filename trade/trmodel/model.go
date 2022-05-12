package trmodel

import (
	"github.com/shopspring/decimal"
)

type OpInfo struct {
	PosNum   int64
	PosPrice decimal.Decimal
	Lim      decimal.Decimal
	Currency string
}
