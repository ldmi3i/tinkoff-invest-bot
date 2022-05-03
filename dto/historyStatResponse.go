package dto

import "github.com/shopspring/decimal"

type HistStatResponse struct {
	BuyOpNum   uint
	SellOpNum  uint
	CurBalance map[string]decimal.Decimal
}
