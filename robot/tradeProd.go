package robot

import (
	"go.uber.org/zap"
	"invest-robot/domain"
	"invest-robot/dto"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy"
	"invest-robot/trade"
)

type TradeProdAPI struct {
	infoSrv    service.InfoSrv
	algFactory strategy.AlgFactory
	algRep     repository.AlgoRepository
	trader     trade.Trader
	logger     *zap.SugaredLogger
}

func (t *TradeProdAPI) Trade(req *dto.CreateAlgorithmRequest) (*dto.TradeStartResponse, error) {
	t.logger.Info("Requested new algorithm ", req)
	algDm := domain.AlgorithmFromDto(req)
	if err := t.algRep.Save(algDm); err != nil {
		return nil, err
	}
	alg, err := t.algFactory.NewProd(algDm)
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
	if err = alg.Go(); err != nil {
		t.logger.Error("Error while starting algorithm, check routine leaking")
		return nil, err
	}
	//TODO check is enough funds for any of requested figis?
	return &dto.TradeStartResponse{Info: "Successfully started"}, nil
}

func NewTradeProdAPI(infoSrv service.InfoSrv, algFactory strategy.AlgFactory, algRep repository.AlgoRepository,
	trader trade.Trader, logger *zap.SugaredLogger) TradeAPI {
	return &TradeProdAPI{infoSrv, algFactory, algRep, trader, logger}
}
