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

func init() {
	algMapping["avr"] = factoryStruct{
		algProd:    avr.NewProd,
		algSandbox: avr.NewSandbox,
		algHist:    avr.NewHist,
	}
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
}

type DefaultAlgFactory struct {
	hRep   repository.HistoryRepository
	iSrv   service.InfoSrv
	cache  map[uint]*stmodel.Algorithm
	logger *zap.SugaredLogger
}

func (a DefaultAlgFactory) NewProd(alg *domain.Algorithm) (stmodel.Algorithm, error) {
	a.logger.Infof("Creating new PROD algorithm with strategy: %s and params: %+v", alg.Strategy, alg.Params)
	factory, exist := algMapping[alg.Strategy]
	if !exist {
		return nil, errors.NewUnexpectedError(
			fmt.Sprintf("Algorithm '%s' does not exist - add mapping to strategy.factory.algMapping", alg.Strategy),
		)
	}
	return factory.algProd(alg, a.iSrv, a.logger)
}

func (a DefaultAlgFactory) NewSandbox(alg *domain.Algorithm) (stmodel.Algorithm, error) {
	a.logger.Infof("Creating new SANDBOX algorithm with strategy: %s and params: %+v", alg.Strategy, alg.Params)
	factory, exist := algMapping[alg.Strategy]
	if !exist {
		return nil, errors.NewUnexpectedError(
			fmt.Sprintf("Algorithm '%s' does not exist - add mapping to strategy.factory.algMapping", alg.Strategy),
		)
	}
	return factory.algSandbox(alg, a.iSrv, a.logger)
}

func (a DefaultAlgFactory) NewHist(alg *domain.Algorithm) (stmodel.Algorithm, error) {
	a.logger.Infof("Creating new history algorithm with strategy: %s and params: %+v", alg.Strategy, alg.Params)
	factory, exist := algMapping[alg.Strategy]
	if !exist {
		return nil, errors.NewUnexpectedError(
			fmt.Sprintf("Algorithm '%s' does not exist - add mapping to strategy.factory.algMapping", alg.Strategy),
		)
	}
	return factory.algHist(alg, a.hRep, a.logger)
}

func NewAlgFactory(srv service.InfoSrv, rep repository.HistoryRepository, logger *zap.SugaredLogger) AlgFactory {
	return DefaultAlgFactory{
		hRep:   rep,
		iSrv:   srv,
		cache:  make(map[uint]*stmodel.Algorithm),
		logger: logger,
	}
}
