package bot

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto/dtotapi"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/errors"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/repository"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/service"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/strategy"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/strategy/stmodel"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/trade"
	"go.uber.org/zap"
)

type TradeAPI interface {
	//Trade starts trading with provided parameters
	Trade(req *dto.CreateAlgorithmRequest, ctx context.Context) (*dto.TradeStartResponse, error)
	//GetActiveAlgorithms returns list of active algorithms
	GetActiveAlgorithms() (*dto.AlgorithmsResponse, error)
	//StopAlgorithm stops algorithm with requested id
	StopAlgorithm(req *dto.StopAlgorithmRequest) (*dto.StopAlgorithmResponse, error)
}

type BaseTradeAPI struct {
	algFactory strategy.AlgFactory
	algRep     repository.AlgoRepository
	trader     trade.Trader
	infoSrv    service.InfoSrv
	logger     *zap.SugaredLogger
}

func (ta *BaseTradeAPI) StopAlgorithm(req *dto.StopAlgorithmRequest) (*dto.StopAlgorithmResponse, error) {
	alg, ok := ta.algFactory.GetAlgorithmById(req.AlgorithmId)
	if !ok {
		return &dto.StopAlgorithmResponse{IsStopped: false, Info: "Algorithm not found"}, nil
	}
	if err := alg.Stop(); err != nil {
		return nil, err
	}
	alg.GetAlgorithm()
	if err := ta.algRep.SetActiveStatus(req.AlgorithmId, false); err != nil {
		ta.logger.Error("Error while setting algorithm to inactive in db! ", err)
	}
	return &dto.StopAlgorithmResponse{IsStopped: true, Info: "Stopped successfully"}, nil
}

func (ta *BaseTradeAPI) tradeInternal(req *dto.CreateAlgorithmRequest,
	factoryF func(request *entity.Algorithm) (stmodel.Algorithm, error), ctx context.Context) (*dto.TradeStartResponse, error) {
	ta.logger.Info("Requested new algorithm ", req)
	//Check is enough rights to account at first
	accounts, err := ta.infoSrv.GetAccounts(ctx)
	if err != nil {
		return nil, err
	}
	acc, ok := accounts.FindAccount(req.AccountId)
	if !ok {
		return nil, errors.NewNotFound("Requested account not found")
	}
	if acc.Status != dtotapi.AccountStatusOpen {
		return nil, errors.NewWrongAccState("Account currently in " + acc.Status.String() + " status")
	}
	if acc.AccessLevel != dtotapi.AccountAccessLevelFullAccess {
		return nil, errors.NewNoAccess("No full access to requested account, check token")
	}

	//Create and start algorithm
	algDm := entity.AlgorithmFromDto(req)
	if err := ta.algRep.Save(algDm); err != nil {
		return nil, err
	}
	alg, err := factoryF(algDm)
	if err != nil {
		return nil, err
	}
	sub, err := alg.Subscribe()
	if err != nil {
		return nil, err
	}
	if err = ta.trader.AddSubscription(sub); err != nil {
		return nil, err
	}
	if err = alg.Go(ctx); err != nil {
		ta.logger.Error("Error while starting algorithm, check routine leaking")
		return nil, err
	}
	return &dto.TradeStartResponse{Info: "Successfully started", AlgorithmID: algDm.ID}, nil
}
