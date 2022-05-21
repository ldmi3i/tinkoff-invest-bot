package robot

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/domain"
	"invest-robot/dto"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy"
	"invest-robot/trade"
)

type TradeSandboxAPI struct {
	*BaseTradeAPI
	infoSrv    service.InfoSrv
	algFactory strategy.AlgFactory
	trader     trade.Trader
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
	t.logger.Info("Requested new algorithm ", req)
	algDm := domain.AlgorithmFromDto(req)
	if err := t.algRep.Save(algDm); err != nil {
		return nil, err
	}
	alg, err := t.algFactory.NewSandbox(algDm)
	if err != nil {
		return nil, err
	}
	sub, err := alg.Subscribe()
	if err != nil {
		return nil, err
	}
	if err = t.trader.AddSubscription(sub); err != nil {
		return nil, err
	}
	if err = alg.Go(ctx); err != nil {
		t.logger.Error("Error while starting algorithm, check routine leaking")
		return nil, err
	}
	return &dto.TradeStartResponse{Info: "Successfully started", AlgorithmID: algDm.ID}, nil
}

func NewSandboxTradeAPI(infoSrv service.InfoSrv, algFactory strategy.AlgFactory, algRep repository.AlgoRepository,
	trader trade.Trader, logger *zap.SugaredLogger) TradeAPI {
	baseAPI := BaseTradeAPI{algFactory: algFactory, logger: logger, algRep: algRep}
	return &TradeSandboxAPI{&baseAPI, infoSrv, algFactory, trader, logger}
}
