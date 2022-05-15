package robot

import (
	"go.uber.org/zap"
	"invest-robot/dto"
	"invest-robot/service"
)

type StatAPI interface {
	GetAlgorithmStat(req *dto.StatAlgoRequest) (*dto.StatAlgoResponse, error)
}

type DefaultStatAPI struct {
	statSrv service.StatService
	logger  *zap.SugaredLogger
}

func NewStatAPI(statRep service.StatService, logger *zap.SugaredLogger) StatAPI {
	return &DefaultStatAPI{statSrv: statRep, logger: logger}
}

func (st *DefaultStatAPI) GetAlgorithmStat(req *dto.StatAlgoRequest) (*dto.StatAlgoResponse, error) {
	return st.statSrv.GetAlgorithmStat(req)
}