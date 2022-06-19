package dtotapi

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"time"
)

type OrderStateResponse struct {
	OrderId           string //Идентификатор заявки.
	Figi              string
	LotsReq           int64                //Lots requested
	LotsExec          int64                //Lots executed
	ExecStatus        PostOrderExcecStatus //Current order status.
	InitPrice         *MoneyValue          //Initial order price. Multiplication of requested lots quantity by price.
	ExecPrice         *MoneyValue          //Executed order price. Multiplication of average price by number of lots.
	TotalPrice        *MoneyValue          //Final order price with all commissions
	AvrPrice          *MoneyValue          //Average single position price.
	InitCommission    *MoneyValue          //Commission calculated at the order initiation
	ExecCommission    *MoneyValue          //Commission calculated by the order result
	Direction         OrderDirection       //Order direction
	InitSecurityPrice *MoneyValue          //Initial price by one instrument
	Stages            []*OrderStage        //Stages of order execution
	ServiceCommission *MoneyValue          //Service commission
	Currency          string
	Type              OrderType
	OrderDate         time.Time //Order creation time
}

type OrderStage struct {
	Price    *MoneyValue
	Quantity int64
	TradeId  string
}

func OrderStageToDto(os *investapi.OrderStage) *OrderStage {
	return &OrderStage{
		Price:    MoneyValueToDto(os.Price),
		Quantity: os.Quantity,
		TradeId:  os.TradeId,
	}
}

func OrderStateResponseToDto(osr *investapi.OrderState) *OrderStateResponse {
	stages := make([]*OrderStage, 0, len(osr.Stages))
	for _, stage := range osr.Stages {
		stages = append(stages, OrderStageToDto(stage))
	}
	return &OrderStateResponse{
		OrderId:           osr.OrderId,
		Figi:              osr.Figi,
		LotsReq:           osr.LotsRequested,
		LotsExec:          osr.LotsExecuted,
		ExecStatus:        PostOrderExcecStatus(osr.ExecutionReportStatus),
		InitPrice:         MoneyValueToDto(osr.InitialOrderPrice),
		ExecPrice:         MoneyValueToDto(osr.ExecutedOrderPrice),
		TotalPrice:        MoneyValueToDto(osr.TotalOrderAmount),
		AvrPrice:          MoneyValueToDto(osr.AveragePositionPrice),
		InitCommission:    MoneyValueToDto(osr.InitialCommission),
		ExecCommission:    MoneyValueToDto(osr.ExecutedCommission),
		Direction:         OrderDirection(osr.Direction),
		InitSecurityPrice: MoneyValueToDto(osr.InitialSecurityPrice),
		Stages:            stages,
		ServiceCommission: MoneyValueToDto(osr.ServiceCommission),
		Currency:          osr.Currency,
		Type:              OrderType(osr.OrderType),
		OrderDate:         osr.OrderDate.AsTime(),
	}
}
