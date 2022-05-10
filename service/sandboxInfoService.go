package service

import (
	"invest-robot/dto/tapi"
	"invest-robot/tinapi"
)

type SandboxInfoService struct {
	*BaseInfoSrv
}

func NewSandboxInfoService(t tinapi.Api) InfoSrv {
	return &SandboxInfoService{newBaseSrv(t)}
}

func (is *SandboxInfoService) GetPositions(req *tapi.PositionsRequest) (*tapi.PositionsResponse, error) {
	return is.tapi.GetSandboxPositions(req)
}

func (is *SandboxInfoService) GetOrderState(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error) {
	return is.tapi.GetSandboxOrderState(req)
}
