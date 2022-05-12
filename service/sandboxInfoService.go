package service

import (
	"go.uber.org/zap"
	"invest-robot/dto/tapi"
	"invest-robot/tinapi"
)

type SandboxInfoService struct {
	*BaseInfoSrv
	logger *zap.SugaredLogger
}

func NewSandboxInfoService(t tinapi.Api, logger *zap.SugaredLogger) InfoSrv {
	return &SandboxInfoService{newBaseSrv(t), logger}
}

func (is *SandboxInfoService) GetPositions(req *tapi.PositionsRequest) (*tapi.PositionsResponse, error) {
	return is.tapi.GetSandboxPositions(req)
}

func (is *SandboxInfoService) GetOrderState(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error) {
	return is.tapi.GetSandboxOrderState(req)
}
