package service

import (
	"context"
	"invest-robot/dto/tapi"
)

type TradeService interface {
	PostOrder(req *tapi.PostOrderRequest, ctx context.Context) (*tapi.PostOrderResponse, error)

	GetOrderStatus(req *tapi.GetOrderStateRequest, ctx context.Context) (*tapi.GetOrderStateResponse, error)

	CancelOrder(req *tapi.CancelOrderRequest, ctx context.Context) (*tapi.CancelOrderResponse, error)
}
