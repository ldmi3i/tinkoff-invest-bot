package robot

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/dto"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy"
	"invest-robot/trade"
)

type TradeProdAPI struct {
	*BaseTradeAPI
	algFactory strategy.AlgFactory
	logger     *zap.SugaredLogger
}

func (t *TradeProdAPI) GetActiveAlgorithms() (*dto.AlgorithmsResponse, error) {
	algs, err := t.algFactory.GetProdAlgs()
	if err != nil {
		t.logger.Error("Error retrieving prod algorithms: ", err)
		return nil, err
	}
	res := make([]*dto.AlgorithmResponse, 0, len(algs))
	for _, alg := range algs {
		res = append(res, alg.GetAlgorithm().ToDto())
	}
	return &dto.AlgorithmsResponse{Algorithms: res}, nil
}

func (t *TradeProdAPI) Trade(req *dto.CreateAlgorithmRequest, ctx context.Context) (*dto.TradeStartResponse, error) {
	return t.tradeInternal(req, t.algFactory.NewProd, ctx)
}

func NewTradeProdAPI(infoSrv service.InfoSrv, algFactory strategy.AlgFactory, algRep repository.AlgoRepository,
	trader trade.Trader, logger *zap.SugaredLogger) TradeAPI {
	baseAPI := BaseTradeAPI{algFactory: algFactory, logger: logger, trader: trader, infoSrv: infoSrv, algRep: algRep}
	return &TradeProdAPI{&baseAPI, algFactory, logger}
}
