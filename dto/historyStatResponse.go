package dto

import (
	"fmt"
	"github.com/shopspring/decimal"
)

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

func (hs HistStatIdDto) String() string {
	return fmt.Sprintf("HistStatIdDto(ID: %d; Param: %s; HistStat: %v)", hs.Id, hs.Param, *hs.HistStat)
}
