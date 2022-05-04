package trmodel

import (
	"github.com/shopspring/decimal"
)

type OpInfo struct {
	LotNum   int64
	LotPrice decimal.Decimal
	Lim      decimal.Decimal
	Currency string
}
