package robot

import (
	"invest-robot/dto"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy"
)

type TradeAPI interface {
	TradeSandbox(req dto.CreateAlgorithmRequest) (*dto.TradeStartResponse, error)
}

type DefaultTradeAPI struct {
	infoSrv service.InfoSrv
	histRep repository.HistoryRepository
	aFact   strategy.AlgFactory
	aRep    repository.AlgoRepository
}

func (t DefaultTradeAPI) TradeSandbox(req dto.CreateAlgorithmRequest) (*dto.TradeStartResponse, error) {

	return nil, errors.NewNotImplemented()
}

func NewSandboxTradeAPI(infoSrv service.InfoSrv, histRep repository.HistoryRepository, aFact strategy.AlgFactory,
	aRep repository.AlgoRepository) TradeAPI {
	return &DefaultTradeAPI{infoSrv, histRep, aFact, aRep}
}
