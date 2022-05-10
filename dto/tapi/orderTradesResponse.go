package tapi

import (
	"github.com/shopspring/decimal"
	"invest-robot/convert"
	investapi "invest-robot/tapigen"
	"time"
)

type OrderTrades struct {
	OrderId   string
	CreatedAt time.Time
	Direction OrderDirection
	Figi      string
	Trades    []*OrderTrade
	AccountId string
}

type OrderTrade struct {
	TradeTime time.Time
	Price     decimal.Decimal
	Quantity  int64
}

func orderTradeToDto(resp *investapi.OrderTrade) *OrderTrade {
	return &OrderTrade{
		TradeTime: resp.DateTime.AsTime(),
		Price:     convert.QuotationToDec(resp.Price),
		Quantity:  resp.Quantity,
	}
}

func OrderTradesToDto(resp *investapi.OrderTrades) *OrderTrades {
	trades := make([]*OrderTrade, 0, len(resp.Trades))
	for _, trade := range resp.Trades {
		trades = append(trades, orderTradeToDto(trade))
	}
	return &OrderTrades{
		OrderId:   resp.OrderId,
		CreatedAt: resp.CreatedAt.AsTime(),
		Direction: OrderDirection(resp.Direction),
		Figi:      resp.Figi,
		Trades:    trades,
		AccountId: resp.AccountId,
	}
}
