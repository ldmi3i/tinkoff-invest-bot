package service

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/dto/tapi"
	"invest-robot/tinapi"
)

type TradeSandboxService struct {
	TinApi tinapi.Api
	logger *zap.SugaredLogger
}

func (ts *TradeSandboxService) CancelOrder(req *tapi.CancelOrderRequest, ctx context.Context) (*tapi.CancelOrderResponse, error) {
	return ts.TinApi.CancelSandboxOrder(req, ctx)
}

func (ts *TradeSandboxService) PostOrder(req *tapi.PostOrderRequest, ctx context.Context) (*tapi.PostOrderResponse, error) {
	return ts.TinApi.PostSandboxOrder(req, ctx)
}

func (ts *TradeSandboxService) GetOrderStatus(req *tapi.GetOrderStateRequest, ctx context.Context) (*tapi.GetOrderStateResponse, error) {
	return ts.TinApi.GetSandboxOrderState(req, ctx)
}

func NewTradeSandboxSrv(tapi tinapi.Api, logger *zap.SugaredLogger) TradeService {
	return &TradeSandboxService{TinApi: tapi, logger: logger}
}
