package dto

import (
	"fmt"
	"github.com/shopspring/decimal"
)

//HistStatResponse statistics retrieved by algorithm analysis
//CurBalance is result balance by the end algorithm simulation.
//If there are not sold instruments, price is taken from the last data and converted to currency
type HistStatResponse struct {
	BuyOpNum   uint                       `json:"buyOpNum"`   //Number of buy operations
	SellOpNum  uint                       `json:"sellOpNum"`  //Number of sell operations
	CurBalance map[string]decimal.Decimal `json:"curBalance"` //Result profit
}

//HistStatIdDto represents auxiliary dto used by range analysis
type HistStatIdDto struct {
	Id       uint
	HistStat *HistStatResponse
	Param    map[string]string
}

func (hs HistStatIdDto) String() string {
	return fmt.Sprintf("HistStatIdDto(ID: %d; Param: %s; HistStat: %v)", hs.Id, hs.Param, *hs.HistStat)
}
