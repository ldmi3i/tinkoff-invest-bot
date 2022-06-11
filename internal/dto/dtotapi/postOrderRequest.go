package dtotapi

import (
	"github.com/shopspring/decimal"
	"invest-robot/internal/convert"
	investapi "invest-robot/internal/tapigen"
)

type OrderDirection int
type OrderType int

const (
	OrderDirectionUnspecified OrderDirection = iota
	OrderDirectionBuy
	OrderDirectionSell
)

const (
	OrderTypeUnspecified OrderType = iota
	OrderTypeLimit
	OrderTypeMarket
)

type PostOrderRequest struct {
	Figi       string
	PosNum     int64
	InstrPrice decimal.Decimal
	Direction  OrderDirection
	AccountId  string
	OrderType  OrderType
	OrderId    string
}

func (pr *PostOrderRequest) ToTinApi() *investapi.PostOrderRequest {
	if pr == nil {
		return nil
	}
	return &investapi.PostOrderRequest{
		Figi:      pr.Figi,
		Quantity:  pr.PosNum,
		Price:     convert.DecToQuotation(pr.InstrPrice),
		Direction: investapi.OrderDirection(pr.Direction),
		AccountId: pr.AccountId,
		OrderType: investapi.OrderType(pr.OrderType),
		OrderId:   pr.OrderId,
	}
}
