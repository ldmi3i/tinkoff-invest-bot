package avr

import (
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy/stmodel"
)

type AlgorithmImpl struct {
	id        uint
	isActive  bool
	dataProc  DataProc
	accountId string
	figis     []string
	limits    []*domain.MoneyLimit
	param     map[string]string
	aChan     chan *stmodel.ActionReq
	arChan    chan *stmodel.ActionResp
	buyPrice  map[string]decimal.Decimal
	limWidth  decimal.Decimal

	logger *zap.SugaredLogger
}

func (a *AlgorithmImpl) GetId() uint {
	return a.id
}

type algoStatus int

const (
	process algoStatus = iota
	waitRes
)

const (
	tradeLimWidth string = "limWidth"
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
	aCh := make(chan *stmodel.ActionReq, 1) //must not block algorithm, so size = 1
	a.aChan = aCh

	arCh := make(chan *stmodel.ActionResp, 1) //must not block trader, so size = 1
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
		a.logger.Infof("Stopping algorithm background; ID: %d", a.id)
	}()
	aDat := AlgoData{
		status:      process,
		prev:        make(map[string]decimal.Decimal),
		instrAmount: make(map[string]int64),
	}
	a.logger.Infof("Starting background algorithm processing; id: %d , strategy: avr , limits: %+v",
		a.id, a.limits)
	for {
		select {
		case resp, ok := <-a.arChan:
			a.logger.Debugf("Receiving response, channel state: %t , response: %+v", ok, *resp.Action)
			if ok {
				err := a.processTraderResp(&aDat, resp)
				if err != nil {
					a.logger.Errorf("Error while trader response processing:\n%s", err)
					return
				}
			} else {
				a.logger.Warn("Trader closed response channel, stopping algorithm...")
				return
			}
		case pDat, ok := <-datCh:
			a.logger.Debugf("Receiving data, channel state: %t", ok)
			if ok {
				a.processData(&aDat, &pDat)
			} else {
				a.logger.Infof("Closed data processor stream, stopping algorithm...")
				return
			}
		}
	}
}

//process response from trade.Trader after pass trading stages
func (a *AlgorithmImpl) processTraderResp(aDat *AlgoData, resp *stmodel.ActionResp) error {
	action := resp.Action
	a.logger.Debug("Processing trader response: ", *resp.Action)
	if resp.Action.Status == domain.SUCCESS {
		iAmount := action.LotAmount
		if action.Direction == domain.SELL {
			iAmount = -iAmount
			//Drops buy price - because the deal has already been completed
			delete(a.buyPrice, action.InstrFigi)
		} else {
			//Checks and update previous buy limit to wait for next sell price no lower than buy price
			price, ok := a.buyPrice[action.InstrFigi]
			if ok {
				//If it's consecutive buy then select maximum price
				a.buyPrice[action.InstrFigi] = decimal.Max(price, action.PositionPrice)
			} else {
				a.buyPrice[action.InstrFigi] = action.PositionPrice
			}
		}
		a.logger.Infof("Incrementing instrument: %s with amount %d", action.InstrFigi, iAmount)
		aDat.instrAmount[action.InstrFigi] = aDat.instrAmount[action.InstrFigi] + iAmount
	} else {
		a.logger.Infof("Operation failed %+v", resp)
	}
	a.logger.Infof("Trader response processed, algo data: %+v", aDat)
	aDat.status = process
	return nil
}

func (a *AlgorithmImpl) processData(aDat *AlgoData, pDat *procData) {
	prevDiff, exists := aDat.prev[pDat.Figi]
	currDiff := pDat.SAV.Sub(pDat.LAV)
	a.logger.Debugf("Difference, current: %s, prev: %s, price: %s", currDiff, prevDiff, pDat.Price)
	aDat.prev[pDat.Figi] = currDiff
	if aDat.status != process {
		a.logger.Debugf("Waiting in status: %d", aDat.status)
		return
	}
	if exists && prevDiff.IsNegative() && currDiff.IsPositive() {
		price, ok := a.buyPrice[pDat.Figi]
		if ok {
			a.logger.Info("Previous buy operation not finished with price: ", price, "; waiting for sell operation...")
			return
		}
		action := domain.Action{
			AlgorithmID: a.id,
			Direction:   domain.BUY,
			InstrFigi:   pDat.Figi,
			Status:      domain.CREATED,
			RetrievedAt: pDat.Time,
			AccountID:   a.accountId,
		}
		a.logger.Infof("Conditions for BUY, requesting action: %+v", action)
		a.aChan <- a.makeReq(&action)
		aDat.status = waitRes
	} else if exists && prevDiff.IsPositive() && currDiff.IsNegative() {
		amount, iExists := aDat.instrAmount[pDat.Figi]
		buyPrice, ok := a.buyPrice[pDat.Figi]
		a.logger.Infof("Check to sell; Instrument: %s; amount: %d", pDat.Figi, amount)
		if ok {
			a.logger.Infof("Buy price found. Current price: %s, buy price: %s", pDat.Price, buyPrice)
			if buyPrice.GreaterThanOrEqual(pDat.Price) {
				a.logger.Info("Buy price is greater than current, discarding sell...")
				return
			}
		}
		if iExists && amount != 0 {
			action := domain.Action{
				AlgorithmID: a.id,
				Direction:   domain.SELL,
				InstrFigi:   pDat.Figi,
				LotAmount:   amount,
				Status:      domain.CREATED,
				RetrievedAt: pDat.Time,
				AccountID:   a.accountId,
			}
			a.logger.Infof("Conditions for SELL, requesting action: %+v", action)
			a.aChan <- a.makeReq(&action)
			aDat.status = waitRes
		}
	}
}

func (a *AlgorithmImpl) makeReq(action *domain.Action) *stmodel.ActionReq {
	return &stmodel.ActionReq{
		Action: action,
		Limits: a.limits,
	}
}

func (a *AlgorithmImpl) Stop() error {
	return errors.NewNotImplemented()
}

func (a *AlgorithmImpl) Configure(ctx []domain.CtxParam) error {
	return errors.NewNotImplemented()
}

func (a *AlgorithmImpl) GetParam() map[string]string {
	return a.param
}

func NewProd(algo *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	proc, err := newProdDataProc(algo, infoSrv, logger)
	if err != nil {
		return nil, err
	}
	paramMap := domain.ParamsToMap(algo.Params)
	var limWidth decimal.Decimal
	limWidthStr, ok := paramMap[tradeLimWidth]
	if ok {
		limWidth, err = decimal.NewFromString(limWidthStr)
		if err != nil {
			logger.Error("Error parsing limit width: ", err)
			return nil, err
		}
	} else {
		limWidth = decimal.NewFromFloat(0.01)
	}
	return &AlgorithmImpl{
		id:        algo.ID,
		isActive:  true,
		accountId: algo.AccountId,
		dataProc:  proc,
		figis:     algo.Figis,
		limits:    algo.MoneyLimits,
		param:     paramMap,
		buyPrice:  make(map[string]decimal.Decimal),
		logger:    logger,
		limWidth:  limWidth,
	}, nil
}

func NewSandbox(algo *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	proc, err := newSandboxDataProc(algo, infoSrv, logger)
	if err != nil {
		return nil, err
	}
	paramMap := domain.ParamsToMap(algo.Params)
	var limWidth decimal.Decimal
	limWidthStr, ok := paramMap[tradeLimWidth]
	if ok {
		limWidth, err = decimal.NewFromString(limWidthStr)
		if err != nil {
			logger.Error("Error parsing limit width: ", err)
			return nil, err
		}
	} else {
		limWidth = decimal.NewFromFloat(0.01)
	}
	return &AlgorithmImpl{
		id:        algo.ID,
		isActive:  true,
		accountId: algo.AccountId,
		dataProc:  proc,
		figis:     algo.Figis,
		limits:    algo.MoneyLimits,
		param:     paramMap,
		buyPrice:  make(map[string]decimal.Decimal),
		logger:    logger,
		limWidth:  limWidth,
	}, nil
}

func NewHist(algo *domain.Algorithm, hRep repository.HistoryRepository, rootLogger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	//New logger with Increased level to suppress history analysis not necessary logging
	logger := zap.New(rootLogger.Desugar().Core(), zap.IncreaseLevel(zap.WarnLevel)).Sugar()
	proc, err := newHistoryDataProc(algo, hRep, logger)
	if err != nil {
		return nil, err
	}
	paramMap := domain.ParamsToMap(algo.Params)
	var limWidth decimal.Decimal
	limWidthStr, ok := paramMap[tradeLimWidth]
	if ok {
		limWidth, err = decimal.NewFromString(limWidthStr)
		if err != nil {
			rootLogger.Error("Error parsing limit width: ", err)
			return nil, err
		}
	} else {
		limWidth = decimal.NewFromFloat(0.01)
	}
	return &AlgorithmImpl{
		id:        algo.ID,
		isActive:  true,
		accountId: algo.AccountId,
		dataProc:  proc,
		figis:     algo.Figis,
		limits:    algo.MoneyLimits,
		param:     paramMap,
		buyPrice:  make(map[string]decimal.Decimal),
		logger:    logger,
		limWidth:  limWidth,
	}, nil
}
