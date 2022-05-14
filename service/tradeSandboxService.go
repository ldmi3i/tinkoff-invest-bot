package service

import (
	"go.uber.org/zap"
	"invest-robot/dto/tapi"
	"invest-robot/tinapi"
)

type TradeSandboxService struct {
	TinApi tinapi.Api
	logger *zap.SugaredLogger
}

func (ts *TradeSandboxService) PostOrder(req *tapi.PostOrderRequest) (*tapi.PostOrderResponse, error) {
	return ts.TinApi.PostSandboxOrder(req)
}

func (ts *TradeSandboxService) GetOrderStatus(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error) {
	return ts.TinApi.GetSandboxOrderState(req)
}

func NewTradeSandboxSrv(tapi tinapi.Api, logger *zap.SugaredLogger) TradeService {
	return &TradeSandboxService{TinApi: tapi, logger: logger}
}
