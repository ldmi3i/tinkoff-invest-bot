package avr

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/tevino/abool/v2"
	"go.uber.org/zap"
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy/stmodel"
	"strconv"
	"time"
)

//AlgorithmImpl is base implementation of average trading algorithm.
//Data processor is aggregation part of algorithm and controlled by the algorithm.
//This implementation supports one subscription and communication with one trader
//Algorithm based on 2 moving average windows - short and long.
//Average window values calculated and provided by DataProc
//Average algorithm just check the difference between windows and compare sign with previous step
//if difference changes from negative to positive - then there is trend for rising and buy condition is met
//if difference changes from positive to negative - then there is trend to falling and sell condition is met
//
//Important note! Currently, at start algorithm assumes that there is no available instruments to sell (it may be added in future as parameter or domain.Algorithm)
//So at start algorithm search for buy conditions and only after that it cat make sell operations
type AlgorithmImpl struct {
	id          uint                       //Algorithm id extracted for more convenience
	isActive    *abool.AtomicBool          //Atomic bool indicating is algorithm active
	dataProc    DataProc                   //Data processor - provides data as the channel for algorithm
	accountId   string                     //Account id extracted for more convenience
	figis       []string                   //List of figis to monitor and use in algorithm
	limits      []*domain.MoneyLimit       //Limits of money available for algorithm
	algorithm   *domain.Algorithm          //Link to original object which algorithm based
	param       map[string]string          //Map of different algorithm configuration parameters (order expiration time etc)
	aChan       chan *stmodel.ActionReq    //Channel to send order requests to trader
	arChan      chan *stmodel.ActionResp   //Channel to receive responses from trader about action result
	stopCh      chan bool                  //Channel to stop algorithm when required
	buyPrice    map[string]decimal.Decimal //Cache of buy prices made previously (when sell goes after buy - it clears record) - to prevent selling cheaper than previous buy
	ordExp      time.Duration              //Expiration duration of posted orders - when expiration time passed and order not finished then it will be canceled
	commission  decimal.Decimal            //Commission on deals to take into account
	ctx         context.Context
	cancelF     context.CancelFunc
	instrAmount map[string]int64 //Initial amount of instruments available

	logger *zap.SugaredLogger
}

type algoStatus int

const (
	process algoStatus = iota
	waitRes
)

const (
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
	return a.isActive.IsSet()
}

func (a *AlgorithmImpl) Go(parCtx context.Context) error {
	a.ctx, a.cancelF = context.WithCancel(parCtx)
	ch, err := a.dataProc.GetDataStream()
	if err != nil {
		return err
	}
	go a.procBg(ch)
	err = a.dataProc.Go(a.ctx)
	if err != nil {
		a.logger.Error("Error while starting data processor: ", err)
		a.stopInternal()
		return errors.NewUnexpectedError("Error while starting data processor " + err.Error())
	}
	a.isActive.Set()
	return nil
}

func (a *AlgorithmImpl) procBg(datCh <-chan procData) {
	defer func() {
		a.isActive.UnSet()
		close(a.aChan)
		a.logger.Infof("Stopping algorithm background; ID: %d", a.id)
	}()
	aDat := AlgoData{ //Algorithm data storing as single thread state
		status:      process,
		prev:        make(map[string]decimal.Decimal),
		instrAmount: a.instrAmount,
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
			if ok {
				a.processData(&aDat, &pDat)
			} else {
				a.logger.Infof("Closed data processor stream, stopping algorithm...")
				return
			}
		case <-a.ctx.Done():
			a.logger.Info("Context canceled, stopping...")
			return
		}
	}
}

//process response from trade.Trader after requested passed trading stages
func (a *AlgorithmImpl) processTraderResp(aDat *AlgoData, resp *stmodel.ActionResp) error {
	action := resp.Action
	a.logger.Debug("Processing trader response: ", *resp.Action)
	if resp.Action.Status == domain.Success {
		iAmount := action.LotAmount
		if action.Direction == domain.Sell {
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
			Direction:      domain.Buy,
			InstrFigi:      pDat.Figi,
			ReqPrice:       pDat.Price,
			ExpirationTime: time.Now().Add(a.ordExp),
			Status:         domain.Created,
			OrderType:      domain.Limited,
			RetrievedAt:    pDat.Time,
			AccountID:      a.accountId,
		}
		a.logger.Infof("Conditions for Buy, requesting action: %+v", action)
		a.aChan <- a.makeReq(&action)
		aDat.status = waitRes
	} else if exists && prevDiff.IsPositive() && currDiff.IsNegative() {
		amount, iExists := aDat.instrAmount[pDat.Figi]
		buyPrice, ok := a.buyPrice[pDat.Figi]
		a.logger.Infof("Check to sell; Instrument: %s; amount: %d", pDat.Figi, amount)
		if ok {
			buyPriceComm := buyPrice.Mul(decimal.NewFromInt(1).Add(a.commission.Mul(decimal.NewFromInt(2))))
			a.logger.Infof("Buy price found. Current price: %s, buy price: %s, buy with percents: %s", pDat.Price, buyPrice, buyPriceComm)
			if buyPriceComm.GreaterThanOrEqual(pDat.Price) {
				a.logger.Infof("Buy price %s is greater than current %s plus 2x commissions - not good enough, waiting better...",
					buyPriceComm, pDat.Price)
				return
			}
		}
		if iExists && amount != 0 {
			action := domain.Action{
				AlgorithmID:    a.id,
				Direction:      domain.Sell,
				InstrFigi:      pDat.Figi,
				LotAmount:      amount,
				ReqPrice:       pDat.Price,
				ExpirationTime: time.Now().Add(a.ordExp),
				Status:         domain.Created,
				OrderType:      domain.Limited,
				RetrievedAt:    pDat.Time,
				AccountID:      a.accountId,
			}
			a.logger.Infof("Conditions for Sell, requesting action: %+v", action)
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
	if a.isActive.IsNotSet() {
		a.logger.Info("Algorithm already stopped, do nothing...")
		return nil
	}
	a.stopInternal()
	return nil
}

func (a *AlgorithmImpl) stopInternal() {
	a.cancelF()
	a.logger.Infof("Algorithm %d successfully stopped", a.algorithm.ID)
}

func (a *AlgorithmImpl) Configure(ctx []*domain.CtxParam) error {
	ctxParam := domain.ContextToMap(ctx)

	return configure(ctxParam, &algoState{
		InitAmount: a.instrAmount,
		BuyPrice:   a.buyPrice,
	}, a.logger)
}

func (a *AlgorithmImpl) GetParam() map[string]string {
	return a.param
}

func (a *AlgorithmImpl) GetAlgorithm() *domain.Algorithm {
	return a.algorithm
}

//NewProd constructs new algorithm using production data processor
func NewProd(algo *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	proc, err := newDataProc(algo, infoSrv, logger)
	if err != nil {
		return nil, err
	}
	return newAvr(algo, logger, proc)
}

//NewSandbox constructs new algorithm using production data processor cause it the same for such algorithm
func NewSandbox(algo *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	proc, err := newDataProc(algo, infoSrv, logger)
	if err != nil {
		return nil, err
	}
	return newAvr(algo, logger, proc)
}

//NewHist constructs new algorithm using history data processor
func NewHist(algo *domain.Algorithm, hRep repository.HistoryRepository, rootLogger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	//New logger with Increased level to suppress history analysis not necessary logging
	logger := zap.New(rootLogger.Desugar().Core(), zap.IncreaseLevel(zap.WarnLevel)).Sugar()
	proc, err := newHistoryDataProc(algo, hRep, logger)
	if err != nil {
		return nil, err
	}
	return newAvr(algo, logger, proc)
}

//Main average algorithm constructor
func newAvr(algo *domain.Algorithm, logger *zap.SugaredLogger, proc DataProc) (stmodel.Algorithm, error) {
	//Turn params to map for convenience
	paramMap := domain.ParamsToMap(algo.Params)
	//Set order expiration time in seconds (when using limited requests), default 5 min
	ordExpInt := getOrDefaultInt(paramMap, OrderExpiration, 300)
	return &AlgorithmImpl{
		id:          algo.ID,
		isActive:    abool.NewBool(true),
		accountId:   algo.AccountId,
		dataProc:    proc,
		figis:       algo.Figis,
		limits:      algo.MoneyLimits,
		param:       paramMap,
		algorithm:   algo,
		buyPrice:    make(map[string]decimal.Decimal),
		stopCh:      make(chan bool),
		logger:      logger,
		ordExp:      time.Duration(ordExpInt) * time.Second,
		commission:  getOrDefaultDecimal(paramMap, Commission, decimal.NewFromFloat(0.04)).Div(decimal.NewFromInt(100)),
		instrAmount: make(map[string]int64),
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
