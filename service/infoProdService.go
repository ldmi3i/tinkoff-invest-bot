package service

import (
	"go.uber.org/zap"
	"invest-robot/dto/tapi"
	"invest-robot/tinapi"
)

type InfoProdService struct {
	*BaseInfoSrv
	logger *zap.SugaredLogger
}

func NewInfoProdService(t tinapi.Api, logger *zap.SugaredLogger) InfoSrv {
	return &InfoProdService{newBaseSrv(t), logger}
}

func (is *InfoProdService) GetPositions(req *tapi.PositionsRequest) (*tapi.PositionsResponse, error) {
	return is.tapi.GetProdPositions(req)
}

func (is *InfoProdService) GetOrderState(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error) {
	return is.tapi.GetProdOrderState(req)
}
