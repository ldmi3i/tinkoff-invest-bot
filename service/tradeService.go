package service

import (
	"invest-robot/dto/tapi"
)

type TradeService interface {
	PostOrder(req *tapi.PostOrderRequest) (*tapi.PostOrderResponse, error)

	GetOrderStatus(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error)

	CancelOrder(req *tapi.CancelOrderRequest) (*tapi.CancelOrderResponse, error)
}
