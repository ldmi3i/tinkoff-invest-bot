package robot

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/internal/dto"
	"invest-robot/internal/repository"
	"invest-robot/internal/service"
	"invest-robot/internal/strategy"
	"invest-robot/internal/trade"
)

type TradeSandboxAPI struct {
	*BaseTradeAPI
	algFactory strategy.AlgFactory
	logger     *zap.SugaredLogger
}

func (t *TradeSandboxAPI) GetActiveAlgorithms() (*dto.AlgorithmsResponse, error) {
	algs, err := t.algFactory.GetSdbxAlgs()
	if err != nil {
		t.logger.Error("Error retrieving sandbox algorithms: ", err)
		return nil, err
	}
	res := make([]*dto.AlgorithmResponse, 0, len(algs))
	for _, alg := range algs {
		res = append(res, alg.GetAlgorithm().ToDto())
	}
	return &dto.AlgorithmsResponse{Algorithms: res}, nil
}

func (t TradeSandboxAPI) Trade(req *dto.CreateAlgorithmRequest, ctx context.Context) (*dto.TradeStartResponse, error) {
	return t.tradeInternal(req, t.algFactory.NewSandbox, ctx)
}

func NewSandboxTradeAPI(infoSrv service.InfoSrv, algFactory strategy.AlgFactory, algRep repository.AlgoRepository,
	trader trade.Trader, logger *zap.SugaredLogger) TradeAPI {
	baseAPI := BaseTradeAPI{algFactory: algFactory, logger: logger, trader: trader, infoSrv: infoSrv, algRep: algRep}
	return &TradeSandboxAPI{&baseAPI, algFactory, logger}
}
