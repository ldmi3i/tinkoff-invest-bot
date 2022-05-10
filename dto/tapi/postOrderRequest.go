package tapi

import (
	"github.com/shopspring/decimal"
	"invest-robot/convert"
	investapi "invest-robot/tapigen"
)

type OrderDirection int
type OrderType int

const (
	ORDER_DIRECTION_UNSPECIFIED OrderDirection = iota
	ORDER_DIRECTION_BUY
	ORDER_DIRECTION_SELL

	ORDER_TYPE_UNSPECIFIED OrderType = iota
	ORDER_TYPE_LIMIT
	ORDER_TYPE_MARKET
)

type PostOrderRequest struct {
	Figi       string
	LotNum     int64
	InstrPrice decimal.Decimal
	Direction  OrderDirection
	AccountId  string
	OrderType  OrderType
	OrderId    string
}

func (pr *PostOrderRequest) ToTinApi() *investapi.PostOrderRequest {
	return &investapi.PostOrderRequest{
		Figi:      pr.Figi,
		Quantity:  pr.LotNum,
		Price:     convert.DecToQuotation(pr.InstrPrice),
		Direction: investapi.OrderDirection(pr.Direction),
		AccountId: pr.AccountId,
		OrderType: investapi.OrderType(pr.OrderType),
		OrderId:   pr.OrderId,
	}
}
