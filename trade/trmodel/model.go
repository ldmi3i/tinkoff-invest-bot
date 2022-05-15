package trmodel

import (
	"github.com/shopspring/decimal"
)

type OpInfo struct {
	PosInLot int64
	PosPrice decimal.Decimal
	Lim      decimal.Decimal
	Currency string
}
