package service

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/dto/dtotapi"
	"invest-robot/tinapi"
)

type InfoSandboxService struct {
	*BaseInfoSrv
	logger *zap.SugaredLogger
}

func NewInfoSandboxService(t tinapi.Api, logger *zap.SugaredLogger) InfoSrv {
	return &InfoSandboxService{newBaseSrv(t), logger}
}

func (is *InfoSandboxService) GetPositions(req *dtotapi.PositionsRequest, ctx context.Context) (*dtotapi.PositionsResponse, error) {
	return is.tapi.GetSandboxPositions(req, ctx)
}

func (is *InfoSandboxService) GetOrderState(req *dtotapi.OrderStateRequest, ctx context.Context) (*dtotapi.OrderStateResponse, error) {
	return is.tapi.GetSandboxOrderState(req, ctx)
}

func (is *InfoSandboxService) GetAccounts(ctx context.Context) (*dtotapi.AccountsResponse, error) {
	return is.tapi.GetSandboxAccounts(ctx)
}
