package strategy

import (
	"fmt"
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy/avr"
	"invest-robot/strategy/model"
	"log"
)

type algProdFunc func(req domain.Algorithm, infoSrv *service.InfoSrv) (model.Algorithm, error)
type algHistFunc func(req domain.Algorithm, rep *repository.HistoryRepository) (model.Algorithm, error)

//Mapping for algorithm creation strategy
var algMapping = make(map[string]factoryStruct)

func init() {
	algMapping["avr"] = factoryStruct{
		algProd: avr.NewProd,
		algHist: avr.NewHist,
	}
}

type factoryStruct struct {
	algProd algProdFunc
	algHist algHistFunc
}

type AlgFactory interface {
	NewApi(alg domain.Algorithm) (model.Algorithm, error)
	NewHist(alg domain.Algorithm) (model.Algorithm, error)
}

type DefaultAlgFactory struct {
	hRep  *repository.HistoryRepository
	iSrv  *service.InfoSrv
	cache map[uint]*model.Algorithm
}

func (a DefaultAlgFactory) NewApi(alg domain.Algorithm) (model.Algorithm, error) {
	log.Printf("Creating new PROD algorithm with strategy: %s and params: %+v", alg.Strategy, alg.Params)
	factory, exist := algMapping[alg.Strategy]
	if !exist {
		return nil, errors.NewUnexpectedError(
			fmt.Sprintf("Algorithm '%s' does not exist - add mapping to strategy.factory.algMapping", alg.Strategy),
		)
	}

	return factory.algProd(alg, a.iSrv)
}

func (a DefaultAlgFactory) NewHist(alg domain.Algorithm) (model.Algorithm, error) {
	log.Printf("Creating new history algorithm with strategy: %s and params: %+v", alg.Strategy, alg.Params)
	factory, exist := algMapping[alg.Strategy]
	if !exist {
		return nil, errors.NewUnexpectedError(
			fmt.Sprintf("Algorithm '%s' does not exist - add mapping to strategy.factory.algMapping", alg.Strategy),
		)
	}
	return factory.algHist(alg, a.hRep)
}

func NewAlgFactory(srv *service.InfoSrv, rep *repository.HistoryRepository) AlgFactory {
	return DefaultAlgFactory{
		hRep:  rep,
		iSrv:  srv,
		cache: make(map[uint]*model.Algorithm),
	}
}
