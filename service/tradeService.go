package service

import (
	"context"
	"invest-robot/dto/dtotapi"
)

type TradeService interface {
	PostOrder(req *dtotapi.PostOrderRequest, ctx context.Context) (*dtotapi.PostOrderResponse, error)

	GetOrderStatus(req *dtotapi.GetOrderStateRequest, ctx context.Context) (*dtotapi.GetOrderStateResponse, error)

	CancelOrder(req *dtotapi.CancelOrderRequest, ctx context.Context) (*dtotapi.CancelOrderResponse, error)
}
