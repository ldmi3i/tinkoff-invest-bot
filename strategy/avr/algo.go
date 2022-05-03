package avr

import (
	"github.com/shopspring/decimal"
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy/model"
	"log"
)

type AlgorithmImpl struct {
	id         uint
	isActive   bool
	dataProc   *DataProc
	figis      []string
	currencies []string
	limits     []decimal.Decimal
	param      map[string]string
	aChan      chan model.ActionReq
	arChan     chan model.ActionResp
}

type algoStatus int

const (
	process algoStatus = iota
	waitRes
)

type AlgoData struct {
	status      algoStatus
	prev        map[string]decimal.Decimal
	instrAmount map[string]int64
}

func (a *AlgorithmImpl) Subscribe() (*model.Subscription, error) {
	if a.aChan != nil || a.arChan != nil {
		return nil, errors.NewDoubleSubErr("Avr algorithm multiple subscription not implemented")
	}
	aCh := make(chan model.ActionReq, 1) //must not block algorithm, so size = 1
	a.aChan = aCh

	arCh := make(chan model.ActionResp, 1) //must not block trader, so size = 1
	a.arChan = arCh
	return &model.Subscription{AlgoID: a.id, AChan: a.aChan, RChan: a.arChan}, nil
}

func (a AlgorithmImpl) IsActive() bool {
	return a.isActive
}

func (a *AlgorithmImpl) Go() error {
	ch, err := (*a.dataProc).GetDataStream()
	if err != nil {
		return err
	}
	go a.procBg(ch)
	a.isActive = true
	return nil
}

func (a *AlgorithmImpl) procBg(datCh <-chan procData) {
	defer func() {
		a.isActive = false
		close(a.arChan)
		close(a.aChan)
	}()
	aDat := AlgoData{
		status: process,
		prev:   make(map[string]decimal.Decimal),
	}
	for {
		select {
		case resp, ok := <-a.arChan:
			if ok {
				err := a.processTraderResp(&aDat, &resp)
				if err != nil {
					log.Printf("Error while trader response processing:\n%s", err)
					return
				}
			} else {
				log.Printf("Error - trader closed response channel, stopping algorithm...")
				return
			}
		case pDat, ok := <-datCh:
			if ok {
				a.processData(&aDat, &pDat)
			} else {
				log.Printf("Closed data processor stream, stopping algorithm...")
			}
		}
	}
}

//process response from trade.Trader after pass trading stages
func (a *AlgorithmImpl) processTraderResp(aDat *AlgoData, resp *model.ActionResp) error {
	action := resp.Action
	if resp.IsSuccess {
		iAmount, exists := aDat.instrAmount[action.InstrFigi]
		if !exists {
			aDat.instrAmount[action.InstrFigi] = iAmount
		}
		if action.Direction == domain.SELL {
			iAmount = -iAmount
		}
		aDat.instrAmount[action.InstrFigi] = aDat.instrAmount[action.InstrFigi] + iAmount
	} else {
		log.Printf("Operation failed %+v", resp)
	}
	aDat.status = process
	return nil
}

func (a *AlgorithmImpl) processData(aDat *AlgoData, pDat *procData) {
	val, exists := aDat.prev[pDat.Figi]
	diff := pDat.LAV.Sub(pDat.SAV)
	if aDat.status == process && exists && val.IsNegative() && diff.IsPositive() {
		action := domain.Action{
			AlgorithmID: a.id,
			Direction:   domain.BUY,
			InstrFigi:   pDat.Figi,
			Status:      domain.CREATED,
			RetrievedAt: pDat.Time,
		}
		a.aChan <- a.makeReq(action)
		aDat.status = waitRes
	} else if aDat.status == process && exists && val.IsPositive() && diff.IsNegative() {
		amount, iExists := aDat.instrAmount[pDat.Figi]
		if iExists && amount != 0 {
			action := domain.Action{
				AlgorithmID: a.id,
				Direction:   domain.SELL,
				InstrFigi:   pDat.Figi,
				InstrAmount: amount,
				Status:      domain.CREATED,
				RetrievedAt: pDat.Time,
			}
			a.aChan <- a.makeReq(action)
		}
	}
	aDat.prev[pDat.Figi] = diff
}

func (a *AlgorithmImpl) makeReq(action domain.Action) model.ActionReq {
	return model.ActionReq{
		Action:     action,
		Currencies: a.currencies,
		Limits:     a.limits,
	}
}

func (a *AlgorithmImpl) Stop() error {
	return errors.NewNotImplemented()
}

func (a *AlgorithmImpl) Configure(ctx []domain.CtxParam) error {
	return errors.NewNotImplemented()
}

func NewProd(algo domain.Algorithm, infoSrv *service.InfoSrv) (model.Algorithm, error) {
	proc, err := newApiDataProc(algo, infoSrv)
	if err != nil {
		return nil, err
	}
	return &AlgorithmImpl{
		id:         algo.ID,
		isActive:   true,
		dataProc:   &proc,
		figis:      algo.Figis,
		currencies: algo.Currencies,
		limits:     algo.Limits,
		param:      domain.ParamsToMap(algo.Params),
	}, nil
}

func NewHist(algo domain.Algorithm, hRep *repository.HistoryRepository) (model.Algorithm, error) {
	proc, err := newHistoryDataProc(algo, hRep)
	if err != nil {
		return nil, err
	}
	return &AlgorithmImpl{
		id:         algo.ID,
		isActive:   true,
		dataProc:   &proc,
		figis:      algo.Figis,
		currencies: algo.Currencies,
		limits:     algo.Limits,
		param:      domain.ParamsToMap(algo.Params),
	}, nil
}
