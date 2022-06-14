package service

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto/dtotapi"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tinapi"
	"go.uber.org/zap"
)

type TradeProdService struct {
	TinApi tinapi.Api
	logger *zap.SugaredLogger
}

func (ts *TradeProdService) CancelOrder(req *dtotapi.CancelOrderRequest, ctx context.Context) (*dtotapi.CancelOrderResponse, error) {
	return ts.TinApi.CancelProdOrder(req, ctx)
}

func (ts *TradeProdService) PostOrder(req *dtotapi.PostOrderRequest, ctx context.Context) (*dtotapi.PostOrderResponse, error) {
	return ts.TinApi.PostProdOrder(req, ctx)
}

func (ts *TradeProdService) GetOrderStatus(req *dtotapi.OrderStateRequest, ctx context.Context) (*dtotapi.OrderStateResponse, error) {
	return ts.TinApi.GetProdOrderState(req, ctx)
}

func NewTradeProdService(tapi tinapi.Api, logger *zap.SugaredLogger) TradeService {
	return &TradeProdService{TinApi: tapi, logger: logger}
}
