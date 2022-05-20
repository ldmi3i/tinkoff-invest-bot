package service

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/dto/tapi"
	"invest-robot/tinapi"
)

type TradeProdService struct {
	TinApi tinapi.Api
	logger *zap.SugaredLogger
}

func (ts *TradeProdService) CancelOrder(req *tapi.CancelOrderRequest, ctx context.Context) (*tapi.CancelOrderResponse, error) {
	return ts.TinApi.CancelProdOrder(req, ctx)
}

func (ts *TradeProdService) PostOrder(req *tapi.PostOrderRequest, ctx context.Context) (*tapi.PostOrderResponse, error) {
	return ts.TinApi.PostProdOrder(req, ctx)
}

func (ts *TradeProdService) GetOrderStatus(req *tapi.GetOrderStateRequest, ctx context.Context) (*tapi.GetOrderStateResponse, error) {
	return ts.TinApi.GetProdOrderState(req, ctx)
}

func NewTradeProdService(tapi tinapi.Api, logger *zap.SugaredLogger) TradeService {
	return &TradeProdService{TinApi: tapi, logger: logger}
}
