package service

import (
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

func (is *InfoSandboxService) GetPositions(req *tapi.PositionsRequest) (*tapi.PositionsResponse, error) {
	return is.tapi.GetSandboxPositions(req)
}

func (is *InfoSandboxService) GetOrderState(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error) {
	return is.tapi.GetSandboxOrderState(req)
}
