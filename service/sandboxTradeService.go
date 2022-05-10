package service

import (
	"invest-robot/dto/tapi"
	"invest-robot/tinapi"
)

type SandboxTradeService struct {
	TinApi tinapi.Api
}

func (ts *SandboxTradeService) PostOrder(req *tapi.PostOrderRequest) (*tapi.PostOrderResponse, error) {
	return ts.TinApi.PostSandboxOrder(req)
}

func (ts *SandboxTradeService) GetOrderStatus(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error) {
	return ts.TinApi.GetSandboxOrderState(req)
}

func NewSandboxTradeSrv(tapi tinapi.Api) TradeService {
	return &SandboxTradeService{TinApi: tapi}
}
