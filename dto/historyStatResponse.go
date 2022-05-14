package dto

import "github.com/shopspring/decimal"

type HistStatResponse struct {
	BuyOpNum   uint                       `json:"buyOpNum"`
	SellOpNum  uint                       `json:"sellOpNum"`
	CurBalance map[string]decimal.Decimal `json:"curBalance"`
}

type HistStatIdDto struct {
	Id       uint
	HistStat *HistStatResponse
	Param    map[string]string
}
