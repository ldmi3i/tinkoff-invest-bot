package trmodel

import (
	"github.com/shopspring/decimal"
)

type OpInfo struct {
	PosNum   int64
	LotPrice decimal.Decimal
	Lim      decimal.Decimal
	Currency string
}
