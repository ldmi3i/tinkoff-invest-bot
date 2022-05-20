package service

import (
	"context"
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

func (is *InfoProdService) GetPositions(req *tapi.PositionsRequest, ctx context.Context) (*tapi.PositionsResponse, error) {
	return is.tapi.GetProdPositions(req, ctx)
}

func (is *InfoProdService) GetOrderState(req *tapi.GetOrderStateRequest, ctx context.Context) (*tapi.GetOrderStateResponse, error) {
	return is.tapi.GetProdOrderState(req, ctx)
}
