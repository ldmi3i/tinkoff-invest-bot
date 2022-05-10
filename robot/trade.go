package robot

import (
	"invest-robot/domain"
	"invest-robot/dto"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy"
	"invest-robot/trade"
	"log"
)

type TradeAPI interface {
	TradeSandbox(req *dto.CreateAlgorithmRequest) (*dto.TradeStartResponse, error)
}

type DefaultTradeAPI struct {
	infoSrv    service.InfoSrv
	algFactory strategy.AlgFactory
	algRep     repository.AlgoRepository
	sdxTrader  trade.Trader
	prodTrader trade.Trader
}

func (t DefaultTradeAPI) TradeSandbox(req *dto.CreateAlgorithmRequest) (*dto.TradeStartResponse, error) {
	log.Println("Requested new algorithm", req)
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
	if err = t.sdxTrader.AddSubscription(sub); err != nil {
		return nil, err
	}
	if err = alg.Go(); err != nil {
		log.Printf("Error while starting algorithm, check routine leaking")
		return nil, err
	}
	//TODO check is enough funds for any of requested figis?
	return &dto.TradeStartResponse{Info: "Successfully started"}, nil
}

func NewSandboxTradeAPI(infoSrv service.InfoSrv, algFactory strategy.AlgFactory, algRep repository.AlgoRepository,
	sdxTrader trade.Trader, prodTrader trade.Trader) TradeAPI {
	return &DefaultTradeAPI{infoSrv, algFactory, algRep, sdxTrader, prodTrader}
}
