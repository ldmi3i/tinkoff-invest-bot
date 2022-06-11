package service

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/internal/dto/dtotapi"
	"invest-robot/internal/tinapi"
)

type TradeSandboxService struct {
	TinApi tinapi.Api
	logger *zap.SugaredLogger
}

func (ts *TradeSandboxService) CancelOrder(req *dtotapi.CancelOrderRequest, ctx context.Context) (*dtotapi.CancelOrderResponse, error) {
	return ts.TinApi.CancelSandboxOrder(req, ctx)
}

func (ts *TradeSandboxService) PostOrder(req *dtotapi.PostOrderRequest, ctx context.Context) (*dtotapi.PostOrderResponse, error) {
	return ts.TinApi.PostSandboxOrder(req, ctx)
}

func (ts *TradeSandboxService) GetOrderStatus(req *dtotapi.OrderStateRequest, ctx context.Context) (*dtotapi.OrderStateResponse, error) {
	return ts.TinApi.GetSandboxOrderState(req, ctx)
}

func NewTradeSandboxSrv(tapi tinapi.Api, logger *zap.SugaredLogger) TradeService {
	return &TradeSandboxService{TinApi: tapi, logger: logger}
}
