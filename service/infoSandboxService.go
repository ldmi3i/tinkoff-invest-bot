package service

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/dto/tapi"
	"invest-robot/tinapi"
)

type InfoSandboxService struct {
	*BaseInfoSrv
	logger *zap.SugaredLogger
}

func NewInfoSandboxService(t tinapi.Api, logger *zap.SugaredLogger) InfoSrv {
	return &InfoSandboxService{newBaseSrv(t), logger}
}

func (is *InfoSandboxService) GetPositions(req *tapi.PositionsRequest, ctx context.Context) (*tapi.PositionsResponse, error) {
	return is.tapi.GetSandboxPositions(req, ctx)
}

func (is *InfoSandboxService) GetOrderState(req *tapi.GetOrderStateRequest, ctx context.Context) (*tapi.GetOrderStateResponse, error) {
	return is.tapi.GetSandboxOrderState(req, ctx)
}
