package service

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/dto/dtotapi"
	"invest-robot/tinapi"
)

type InfoProdService struct {
	*BaseInfoSrv
	logger *zap.SugaredLogger
}

func NewInfoProdService(t tinapi.Api, logger *zap.SugaredLogger) InfoSrv {
	return &InfoProdService{newBaseSrv(t), logger}
}

func (is *InfoProdService) GetPositions(req *dtotapi.PositionsRequest, ctx context.Context) (*dtotapi.PositionsResponse, error) {
	return is.tapi.GetProdPositions(req, ctx)
}

func (is *InfoProdService) GetOrderState(req *dtotapi.OrderStateRequest, ctx context.Context) (*dtotapi.OrderStateResponse, error) {
	return is.tapi.GetProdOrderState(req, ctx)
}

func (is *InfoProdService) GetAccounts(ctx context.Context) (*dtotapi.AccountsResponse, error) {
	return is.tapi.GetProdAccounts(ctx)
}
