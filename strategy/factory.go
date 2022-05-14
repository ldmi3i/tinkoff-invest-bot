package strategy

import (
	"fmt"
	"go.uber.org/zap"
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy/avr"
	"invest-robot/strategy/stmodel"
)

type algProdFunc func(req *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error)
type algSandboxFunc func(req *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error)
type algHistFunc func(req *domain.Algorithm, rep repository.HistoryRepository, logger *zap.SugaredLogger) (stmodel.Algorithm, error)

//Mapping for algorithm creation strategy
var algMapping = make(map[string]factoryStruct)

//Mapping for parameter splitter for algorithm
var parSplMapping = make(map[string]stmodel.ParamSplitter)

func initialize(logger *zap.SugaredLogger) {
	algMapping["avr"] = factoryStruct{
		algProd:    avr.NewProd,
		algSandbox: avr.NewSandbox,
		algHist:    avr.NewHist,
	}
	parSplMapping["avr"] = avr.NewParamSplitter(logger)
}

type factoryStruct struct {
	algProd    algProdFunc
	algHist    algHistFunc
	algSandbox algSandboxFunc
}

type AlgFactory interface {
	NewProd(alg *domain.Algorithm) (stmodel.Algorithm, error)
	NewSandbox(alg *domain.Algorithm) (stmodel.Algorithm, error)
	NewHist(alg *domain.Algorithm) (stmodel.Algorithm, error)
	NewRange(alg *domain.Algorithm) ([]stmodel.Algorithm, error)
}

type DefaultAlgFactory struct {
	hRep        repository.HistoryRepository
	infoSdxSrv  service.InfoSrv
	infoProdSrv service.InfoSrv
	cache       map[uint]*stmodel.Algorithm
	logger      *zap.SugaredLogger
}

func (a *DefaultAlgFactory) NewProd(alg *domain.Algorithm) (stmodel.Algorithm, error) {
	a.logger.Infof("Creating new PROD algorithm with strategy: %s and params: %+v", alg.Strategy, alg.Params)
	factory, exist := algMapping[alg.Strategy]
	if !exist {
		return nil, errors.NewUnexpectedError(
			fmt.Sprintf("Algorithm '%s' does not exist - add mapping to strategy.factory.algMapping", alg.Strategy),
		)
	}
	return factory.algProd(alg, a.infoSdxSrv, a.logger)
}

func (a *DefaultAlgFactory) NewSandbox(alg *domain.Algorithm) (stmodel.Algorithm, error) {
	a.logger.Infof("Creating new SANDBOX algorithm with strategy: %s and params: %+v", alg.Strategy, alg.Params)
	factory, exist := algMapping[alg.Strategy]
	if !exist {
		return nil, errors.NewUnexpectedError(
			fmt.Sprintf("Algorithm '%s' does not exist - add mapping to strategy.factory.algMapping", alg.Strategy),
		)
	}
	return factory.algSandbox(alg, a.infoSdxSrv, a.logger)
}

func (a *DefaultAlgFactory) NewHist(alg *domain.Algorithm) (stmodel.Algorithm, error) {
	a.logger.Infof("Creating new history algorithm with strategy: %s , id: %d", alg.Strategy, alg.ID)
	factory, exist := algMapping[alg.Strategy]
	if !exist {
		return nil, errors.NewUnexpectedError(
			fmt.Sprintf("Algorithm '%s' does not exist - add mapping to strategy.factory.algMapping", alg.Strategy),
		)
	}
	return factory.algHist(alg, a.hRep, a.logger)
}

// NewRange Generates range of algorithms working on history data
func (a *DefaultAlgFactory) NewRange(alg *domain.Algorithm) ([]stmodel.Algorithm, error) {
	a.logger.Infof("Split algo with strategy: %s with params: %+v", alg.Strategy, alg.Params)
	splitter, ok := parSplMapping[alg.Strategy]
	if !ok {
		a.logger.Errorf("Splitter for strategy: %s not found", alg.Strategy)
		return nil, errors.NewUnexpectedError("Splitter not found")
	}
	parMap := domain.ParamsToMap(alg.Params)
	split, err := splitter.ParseAndSplit(parMap)
	if err != nil {
		return nil, err
	}
	algoRange := make([]stmodel.Algorithm, 0, len(split))
	for id, param := range split {
		currAlg := alg.CopyNoParam()
		currAlg.ID = uint(id)
		paramStruct := make([]*domain.Param, 0, len(param))
		for key, value := range param {
			paramStruct = append(paramStruct, &domain.Param{Key: key, Value: value})
		}
		currAlg.Params = paramStruct
		algo, err := a.NewHist(currAlg)

		if err != nil {
			return nil, err
		}
		algoRange = append(algoRange, algo)
	}
	return algoRange, nil
}

func NewAlgFactory(infoSdxSrv service.InfoSrv, infoProdSrv service.InfoSrv, rep repository.HistoryRepository,
	logger *zap.SugaredLogger) AlgFactory {
	initialize(logger)

	return &DefaultAlgFactory{
		hRep:        rep,
		infoSdxSrv:  infoSdxSrv,
		infoProdSrv: infoProdSrv,
		cache:       make(map[uint]*stmodel.Algorithm),
		logger:      logger,
	}
}
