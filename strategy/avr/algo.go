package avr

import (
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy/stmodel"
	"strconv"
	"time"
)

type AlgorithmImpl struct {
	id         uint
	isActive   bool
	dataProc   DataProc
	accountId  string
	figis      []string
	limits     []*domain.MoneyLimit
	algorithm  *domain.Algorithm
	param      map[string]string
	aChan      chan *stmodel.ActionReq
	arChan     chan *stmodel.ActionResp
	stopCh     chan bool
	buyPrice   map[string]decimal.Decimal
	limWidth   decimal.Decimal
	ordExp     time.Duration
	commission decimal.Decimal

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
	TradeLimWidth   string = "limWidth"
	OrderExpiration string = "order_expiration"
	Commission      string = "order_commission"
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
		case <-a.stopCh:
			a.logger.Info("Stop signal received, stopping bg task algorithm...")
			return
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
			AlgorithmID:    a.id,
			Direction:      domain.BUY,
			InstrFigi:      pDat.Figi,
			ReqPrice:       pDat.Price.Mul(decimal.NewFromInt(1).Sub(a.commission)),
			ExpirationTime: time.Now().Add(a.ordExp),
			Status:         domain.CREATED,
			OrderType:      domain.LIMITED,
			RetrievedAt:    pDat.Time,
			AccountID:      a.accountId,
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
				AlgorithmID:    a.id,
				Direction:      domain.SELL,
				InstrFigi:      pDat.Figi,
				LotAmount:      amount,
				ReqPrice:       pDat.Price.Mul(decimal.NewFromInt(1).Sub(a.commission)),
				ExpirationTime: time.Now().Add(a.ordExp),
				Status:         domain.CREATED,
				OrderType:      domain.LIMITED,
				RetrievedAt:    pDat.Time,
				AccountID:      a.accountId,
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
	if !a.isActive {
		a.logger.Info("Algorithm already stopped, do nothing...")
		return nil
	}
	a.logger.Info("Stopping algorithm: ", a.algorithm)
	a.stopCh <- true
	err := a.dataProc.Stop()
	if err != nil {
		a.logger.Error("Algorithm stopped, but data processor exit with error!")
		return err
	}
	a.logger.Infof("Algorithm %d successfully stopped", a.algorithm.ID)
	return nil
}

func (a *AlgorithmImpl) Configure(ctx []domain.CtxParam) error {
	return errors.NewNotImplemented()
}

func (a *AlgorithmImpl) GetParam() map[string]string {
	return a.param
}

func (a *AlgorithmImpl) GetAlgorithm() *domain.Algorithm {
	return a.algorithm
}

func (a *AlgorithmImpl) GetLimits() []*domain.MoneyLimit {
	return a.limits
}

func NewProd(algo *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	proc, err := newDataProc(algo, infoSrv, logger)
	if err != nil {
		return nil, err
	}
	paramMap := domain.ParamsToMap(algo.Params)
	var limWidth decimal.Decimal
	limWidthStr, ok := paramMap[TradeLimWidth]
	if ok {
		limWidth, err = decimal.NewFromString(limWidthStr)
		if err != nil {
			logger.Error("Error parsing limit width: ", err)
			return nil, err
		}
	} else {
		limWidth = decimal.NewFromFloat(0.01)
	}
	ordExpInt := getOrDefaultInt(paramMap, OrderExpiration, 300)
	return &AlgorithmImpl{
		id:         algo.ID,
		isActive:   true,
		accountId:  algo.AccountId,
		dataProc:   proc,
		figis:      algo.Figis,
		limits:     algo.MoneyLimits,
		param:      paramMap,
		algorithm:  algo,
		buyPrice:   make(map[string]decimal.Decimal),
		stopCh:     make(chan bool),
		logger:     logger,
		limWidth:   limWidth,
		ordExp:     time.Duration(ordExpInt) * time.Second,
		commission: getOrDefaultDecimal(paramMap, Commission, decimal.NewFromFloat(0.05)).Div(decimal.NewFromInt(100)),
	}, nil
}

func NewSandbox(algo *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	proc, err := newDataProc(algo, infoSrv, logger)
	if err != nil {
		return nil, err
	}
	paramMap := domain.ParamsToMap(algo.Params)
	var limWidth decimal.Decimal
	limWidthStr, ok := paramMap[TradeLimWidth]
	if ok {
		limWidth, err = decimal.NewFromString(limWidthStr)
		if err != nil {
			logger.Error("Error parsing limit width: ", err)
			return nil, err
		}
	} else {
		limWidth = decimal.NewFromFloat(0.01)
	}
	ordExpInt := getOrDefaultInt(paramMap, OrderExpiration, 300)
	return &AlgorithmImpl{
		id:         algo.ID,
		isActive:   true,
		accountId:  algo.AccountId,
		dataProc:   proc,
		figis:      algo.Figis,
		limits:     algo.MoneyLimits,
		param:      paramMap,
		algorithm:  algo,
		buyPrice:   make(map[string]decimal.Decimal),
		stopCh:     make(chan bool),
		logger:     logger,
		limWidth:   limWidth,
		ordExp:     time.Duration(ordExpInt) * time.Second,
		commission: getOrDefaultDecimal(paramMap, Commission, decimal.NewFromFloat(0.05)).Div(decimal.NewFromInt(100)),
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
	limWidthStr, ok := paramMap[TradeLimWidth]
	if ok {
		limWidth, err = decimal.NewFromString(limWidthStr)
		if err != nil {
			rootLogger.Error("Error parsing limit width: ", err)
			return nil, err
		}
	} else {
		limWidth = decimal.NewFromFloat(0.01)
	}
	ordExpInt := getOrDefaultInt(paramMap, OrderExpiration, 300)
	return &AlgorithmImpl{
		id:         algo.ID,
		isActive:   true,
		accountId:  algo.AccountId,
		dataProc:   proc,
		figis:      algo.Figis,
		limits:     algo.MoneyLimits,
		param:      paramMap,
		algorithm:  algo,
		buyPrice:   make(map[string]decimal.Decimal),
		stopCh:     make(chan bool),
		logger:     logger,
		limWidth:   limWidth,
		ordExp:     time.Duration(ordExpInt) * time.Second,
		commission: getOrDefaultDecimal(paramMap, Commission, decimal.NewFromFloat(0.05)).Div(decimal.NewFromInt(100)),
	}, nil
}

func getOrDefaultDecimal(paramMap map[string]string, param string, def decimal.Decimal) decimal.Decimal {
	res, ok := paramMap[param]
	if !ok {
		return def
	}
	resDec, err := decimal.NewFromString(res)
	if err != nil {
		return def
	} else {
		return resDec
	}
}

func getOrDefaultInt(paramMap map[string]string, param string, def int) int {
	res, ok := paramMap[param]
	if !ok {
		return def
	}
	resDec, err := strconv.Atoi(res)
	if err != nil {
		return def
	} else {
		return resDec
	}
}
