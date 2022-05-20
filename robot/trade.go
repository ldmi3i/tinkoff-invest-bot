package robot

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/dto"
	"invest-robot/strategy"
)

type TradeAPI interface {
	Trade(req *dto.CreateAlgorithmRequest, ctx context.Context) (*dto.TradeStartResponse, error)
	GetActiveAlgorithms() (*dto.AlgorithmsResponse, error)
	StopAlgorithm(req *dto.StopAlgorithmRequest) (*dto.StopAlgorithmResponse, error)
}

type BaseTradeAPI struct {
	algFactory strategy.AlgFactory
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
	return &dto.StopAlgorithmResponse{IsStopped: true}, nil
}
