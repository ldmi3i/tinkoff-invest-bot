package dtotapi

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/convert"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"github.com/shopspring/decimal"
)

type PostOrderExcecStatus int

const (
	ExecutionReportStatusUnspecified PostOrderExcecStatus = iota
	ExecutionReportStatusFill
	ExecutionReportStatusRejected
	ExecutionReportStatusCancelled
	ExecutionReportStatusNew
	ExecutionReportStatusPartiallyfill
)

type MoneyValue struct {
	Currency string
	Value    decimal.Decimal
}

func MoneyValueToDto(q *investapi.MoneyValue) *MoneyValue {
	if q == nil {
		return nil
	}
	return &MoneyValue{
		Currency: q.Currency,
		Value:    convert.TinToDec(q.Units, q.Nano),
	}
}

func (mv *MoneyValue) ToTinApi() *investapi.MoneyValue {
	units, nano := convert.DecToTin(mv.Value)
	return &investapi.MoneyValue{
		Currency: mv.Currency,
		Units:    units,
		Nano:     nano,
	}
}

type PostOrderResponse struct {
	OrderId        string
	Figi           string
	LotsReq        int64
	LotsExec       int64
	ExecStatus     PostOrderExcecStatus
	InitPrice      *MoneyValue
	ExecPrice      *MoneyValue
	TotalPrice     *MoneyValue
	InitCommission *MoneyValue
	ExecCommission *MoneyValue
	Aci            *MoneyValue
	Direction      OrderDirection
	Type           OrderType
	Message        string
	InitiOrdPrice  decimal.Decimal
}

func PostOrderResponseToDto(resp *investapi.PostOrderResponse) *PostOrderResponse {
	return &PostOrderResponse{
		OrderId:        resp.OrderId,
		Figi:           resp.Figi,
		LotsReq:        resp.LotsRequested,
		LotsExec:       resp.LotsExecuted,
		ExecStatus:     PostOrderExcecStatus(resp.ExecutionReportStatus),
		InitPrice:      MoneyValueToDto(resp.InitialOrderPrice),
		ExecPrice:      MoneyValueToDto(resp.ExecutedOrderPrice),
		TotalPrice:     MoneyValueToDto(resp.TotalOrderAmount),
		InitCommission: MoneyValueToDto(resp.InitialCommission),
		ExecCommission: MoneyValueToDto(resp.ExecutedCommission),
		Aci:            MoneyValueToDto(resp.AciValue),
		Direction:      OrderDirection(resp.Direction),
		Type:           OrderType(resp.OrderType),
		Message:        resp.Message,
		InitiOrdPrice:  convert.QuotationToDec(resp.InitialOrderPricePt),
	}
}
