package service

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto/dtotapi"
)

//TradeService provides methods for order management
type TradeService interface {
	//PostOrder posts order with requested params
	PostOrder(req *dtotapi.PostOrderRequest, ctx context.Context) (*dtotapi.PostOrderResponse, error)

	//GetOrderStatus returns status and information about the order
	GetOrderStatus(req *dtotapi.OrderStateRequest, ctx context.Context) (*dtotapi.OrderStateResponse, error)

	//CancelOrder cancels order and returns cancel result
	CancelOrder(req *dtotapi.CancelOrderRequest, ctx context.Context) (*dtotapi.CancelOrderResponse, error)
}
