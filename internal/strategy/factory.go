package strategy

import (
	"fmt"
	"go.uber.org/zap"
	"invest-robot/internal/collections"
	"invest-robot/internal/domain"
	"invest-robot/internal/errors"
	"invest-robot/internal/repository"
	"invest-robot/internal/service"
	"invest-robot/internal/strategy/avr"
	"invest-robot/internal/strategy/stmodel"
)

//algProdFunc represents common production algorithm factory method
type algProdFunc func(req *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error)

//algSandboxFunc represents common sandbox algorithm factory method
type algSandboxFunc func(req *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error)

//algHistFunc represents common historical algorithm factory method
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

//AlgFactory provides methods to create new algorithms for different environments
//Also factory caches created algorithms and provide methods with active one
type AlgFactory interface {
	//NewProd returns new production algorithm based on provided properties
	NewProd(alg *domain.Algorithm) (stmodel.Algorithm, error)
	//NewSandbox returns new sandbox algorithm based on provided properties
	NewSandbox(alg *domain.Algorithm) (stmodel.Algorithm, error)
	//NewHist returns algorithm for simulation on historical data
	NewHist(alg *domain.Algorithm) (stmodel.Algorithm, error)
	//NewRange returns slice of algorithms from provided range for simulation on historical data
	NewRange(alg *domain.Algorithm) ([]stmodel.Algorithm, error)
	//GetProdAlgs returns active algorithms running production environment
	GetProdAlgs() ([]stmodel.Algorithm, error)
	//GetSdbxAlgs returns active algorithms running sandbox environment
	GetSdbxAlgs() ([]stmodel.Algorithm, error)
	//GetAlgorithmById returns active algorithm by id, searches sandbox and prod environment
	GetAlgorithmById(algoId uint) (stmodel.Algorithm, bool)
}

type DefaultAlgFactory struct {
	hRep           repository.HistoryRepository
	infoSdxSrv     service.InfoSrv
	infoProdSrv    service.InfoSrv
	prodAlgorithms collections.SyncMap[uint, stmodel.Algorithm]
	sdbxAlgorithms collections.SyncMap[uint, stmodel.Algorithm]
	logger         *zap.SugaredLogger
}

func (a *DefaultAlgFactory) GetAlgorithmById(algoId uint) (stmodel.Algorithm, bool) {
	res, ok := a.prodAlgorithms.Get(algoId)
	if ok {
		return res, true
	} else {
		res, ok = a.sdbxAlgorithms.Get(algoId)
		if ok {
			return res, true
		} else {
			a.logger.Infof("Algorithm by id %d not found", algoId)
			return nil, false
		}
	}
}

func (a *DefaultAlgFactory) GetProdAlgs() ([]stmodel.Algorithm, error) {
	res := make([]stmodel.Algorithm, 0)
	for _, entry := range a.prodAlgorithms.GetSlice() {
		if entry.Value.IsActive() {
			res = append(res, entry.Value)
		}
	}
	return res, nil
}

func (a *DefaultAlgFactory) GetSdbxAlgs() ([]stmodel.Algorithm, error) {
	res := make([]stmodel.Algorithm, 0)
	for _, entry := range a.sdbxAlgorithms.GetSlice() {
		if entry.Value.IsActive() {
			res = append(res, entry.Value)
		}
	}
	return res, nil
}

func (a *DefaultAlgFactory) NewProd(alg *domain.Algorithm) (stmodel.Algorithm, error) {
	a.logger.Infof("Creating new PROD algorithm with strategy: %s and params: %+v", alg.Strategy, alg.Params)
	factory, exist := algMapping[alg.Strategy]
	if !exist {
		return nil, errors.NewUnexpectedError(
			fmt.Sprintf("Algorithm '%s' does not exist - add mapping to strategy.factory.algMapping", alg.Strategy),
		)
	}
	res, err := factory.algProd(alg, a.infoSdxSrv, a.logger)
	if err == nil {
		a.prodAlgorithms.Put(alg.ID, res)
	}
	return res, err
}

func (a *DefaultAlgFactory) NewSandbox(alg *domain.Algorithm) (stmodel.Algorithm, error) {
	a.logger.Infof("Creating new SANDBOX algorithm with strategy: %s and params: %+v", alg.Strategy, alg.Params)
	factory, exist := algMapping[alg.Strategy]
	if !exist {
		a.logger.Error("No mapping for requested algorithm strategy ", alg.Strategy)
		return nil, errors.NewUnexpectedError(
			fmt.Sprintf("Algorithm '%s' does not exist - add mapping to strategy.factory.algMapping", alg.Strategy),
		)
	}
	res, err := factory.algSandbox(alg, a.infoSdxSrv, a.logger)
	if err == nil {
		a.sdbxAlgorithms.Put(alg.ID, res)
	}
	return res, err
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
		hRep:           rep,
		infoSdxSrv:     infoSdxSrv,
		infoProdSrv:    infoProdSrv,
		prodAlgorithms: collections.NewSyncMap[uint, stmodel.Algorithm](),
		sdbxAlgorithms: collections.NewSyncMap[uint, stmodel.Algorithm](),
		logger:         logger,
	}
}