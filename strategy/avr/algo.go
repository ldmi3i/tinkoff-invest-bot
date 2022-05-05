package avr

import (
	"github.com/shopspring/decimal"
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy/stmodel"
	"log"
)

type AlgorithmImpl struct {
	id         uint
	isActive   bool
	dataProc   DataProc
	figis      []string
	currencies []string
	limits     []decimal.Decimal
	param      map[string]string
	aChan      chan stmodel.ActionReq
	arChan     chan stmodel.ActionResp
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

func (a *AlgorithmImpl) Subscribe() (*stmodel.Subscription, error) {
	if a.aChan != nil || a.arChan != nil {
		return nil, errors.NewDoubleSubErr("Avr algorithm multiple subscription not implemented")
	}
	aCh := make(chan stmodel.ActionReq, 1) //must not block algorithm, so size = 1
	a.aChan = aCh

	arCh := make(chan stmodel.ActionResp, 1) //must not block trader, so size = 1
	a.arChan = arCh
	return &stmodel.Subscription{AlgoID: a.id, AChan: a.aChan, RChan: a.arChan}, nil
}

func (a AlgorithmImpl) IsActive() bool {
	return a.isActive
}

func (a *AlgorithmImpl) Go() error {
	ch, err := a.dataProc.GetDataStream()
	if err != nil {
		return err
	}
	go a.procBg(ch)
	a.dataProc.Go()
	a.isActive = true
	return nil
}

func (a *AlgorithmImpl) procBg(datCh <-chan procData) {
	defer func() {
		a.isActive = false
		close(a.arChan)
		close(a.aChan)
		log.Printf("Stopping algorithm background; ID: %d", a.id)
	}()
	aDat := AlgoData{
		status:      process,
		prev:        make(map[string]decimal.Decimal),
		instrAmount: make(map[string]int64),
	}
	log.Printf("Starting background algorithm processing; id: %d , strategy: avr , currencies: %s , limits: %s",
		a.id, a.currencies, a.limits)
	for {
		select {
		case resp, ok := <-a.arChan:
			log.Printf("Receiving response, channel state: %t , response: %+v", ok, resp)
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
			log.Printf("Receiving data, channel state: %t", ok)
			if ok {
				a.processData(&aDat, &pDat)
			} else {
				log.Printf("Closed data processor stream, stopping algorithm...")
				return
			}
		}
	}
}

//process response from trade.Trader after pass trading stages
func (a *AlgorithmImpl) processTraderResp(aDat *AlgoData, resp *stmodel.ActionResp) error {
	action := resp.Action
	if resp.Action.Status == domain.SUCCESS {
		iAmount := action.InstrAmount
		if action.Direction == domain.SELL {
			iAmount = -iAmount
		}
		log.Printf("Incrementing instrument: %s with amount %d", action.InstrFigi, iAmount)
		aDat.instrAmount[action.InstrFigi] = aDat.instrAmount[action.InstrFigi] + iAmount
	} else {
		log.Printf("Operation failed %+v", resp)
	}
	log.Printf("Trader response processed, algo data: %+v", aDat)
	aDat.status = process
	return nil
}

func (a *AlgorithmImpl) processData(aDat *AlgoData, pDat *procData) {
	prevDiff, exists := aDat.prev[pDat.Figi]
	currDiff := pDat.SAV.Sub(pDat.LAV)
	log.Printf("Difference, current: %s, prev: %s", currDiff, prevDiff)
	aDat.prev[pDat.Figi] = currDiff
	if aDat.status != process {
		log.Printf("Waiting in status: %d", aDat.status)
		return
	}
	if exists && prevDiff.IsNegative() && currDiff.IsPositive() {
		action := domain.Action{
			AlgorithmID: a.id,
			Direction:   domain.BUY,
			InstrFigi:   pDat.Figi,
			Status:      domain.CREATED,
			RetrievedAt: pDat.Time,
		}
		log.Printf("Conditions for BUY, requesting action: %+v", action)
		a.aChan <- a.makeReq(action)
		aDat.status = waitRes
	} else if exists && prevDiff.IsPositive() && currDiff.IsNegative() {

		amount, iExists := aDat.instrAmount[pDat.Figi]
		log.Printf("Check to sell; Instrument: %s; amount: %d", pDat.Figi, amount)
		if iExists && amount != 0 {
			action := domain.Action{
				AlgorithmID: a.id,
				Direction:   domain.SELL,
				InstrFigi:   pDat.Figi,
				InstrAmount: amount,
				Status:      domain.CREATED,
				RetrievedAt: pDat.Time,
			}
			log.Printf("Conditions for SELL, requesting action: %+v", action)
			a.aChan <- a.makeReq(action)
			aDat.status = waitRes
		}
	}
}

func (a *AlgorithmImpl) makeReq(action domain.Action) stmodel.ActionReq {
	return stmodel.ActionReq{
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

func NewProd(algo *domain.Algorithm, infoSrv service.InfoSrv) (stmodel.Algorithm, error) {
	proc, err := newProdDataProc(algo, infoSrv)
	if err != nil {
		return nil, err
	}
	return &AlgorithmImpl{
		id:         algo.ID,
		isActive:   true,
		dataProc:   proc,
		figis:      algo.Figis,
		currencies: algo.Currencies,
		limits:     algo.Limits,
		param:      domain.ParamsToMap(algo.Params),
	}, nil
}

func NewSandbox(algo *domain.Algorithm, infoSrv service.InfoSrv) (stmodel.Algorithm, error) {
	proc, err := newSandboxDataProc(algo, infoSrv)
	if err != nil {
		return nil, err
	}
	return &AlgorithmImpl{
		id:         algo.ID,
		isActive:   true,
		dataProc:   proc,
		figis:      algo.Figis,
		currencies: algo.Currencies,
		limits:     algo.Limits,
		param:      domain.ParamsToMap(algo.Params),
	}, nil
}

func NewHist(algo *domain.Algorithm, hRep repository.HistoryRepository) (stmodel.Algorithm, error) {
	proc, err := newHistoryDataProc(algo, hRep)
	if err != nil {
		return nil, err
	}
	return &AlgorithmImpl{
		id:         algo.ID,
		isActive:   true,
		dataProc:   proc,
		figis:      algo.Figis,
		currencies: algo.Currencies,
		limits:     algo.Limits,
		param:      domain.ParamsToMap(algo.Params),
	}, nil
}
