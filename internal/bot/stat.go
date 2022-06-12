package bot

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/service"
	"go.uber.org/zap"
)

type StatAPI interface {
	//GetAlgorithmStat collects and returns statistics on the algorithm with requested id
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
