package service

import (
	"go.uber.org/zap"
	"invest-robot/dto/tapi"
	"invest-robot/tinapi"
)

type TradeProdService struct {
	TinApi tinapi.Api
	logger *zap.SugaredLogger
}

func (ts *TradeProdService) CancelOrder(req *tapi.CancelOrderRequest) (*tapi.CancelOrderResponse, error) {
	return ts.TinApi.CancelProdOrder(req)
}

func (ts *TradeProdService) PostOrder(req *tapi.PostOrderRequest) (*tapi.PostOrderResponse, error) {
	return ts.TinApi.PostProdOrder(req)
}

func (ts *TradeProdService) GetOrderStatus(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error) {
	return ts.TinApi.GetProdOrderState(req)
}

func NewTradeProdService(tapi tinapi.Api, logger *zap.SugaredLogger) TradeService {
	return &TradeProdService{TinApi: tapi, logger: logger}
}
