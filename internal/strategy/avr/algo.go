package avr

import (
	"context"
	"encoding/json"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/errors"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/repository"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/service"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/strategy/stmodel"
	"github.com/shopspring/decimal"
	"github.com/tevino/abool/v2"
	"go.uber.org/zap"
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
	id              uint                       //Algorithm id extracted for more convenience
	isActive        *abool.AtomicBool          //Atomic bool indicating is algorithm active
	dataProc        DataProc                   //Data processor - provides data as the channel for algorithm
	accountId       string                     //Account id extracted for more convenience
	figis           []string                   //List of figis to monitor and use in algorithm
	limits          []*entity.MoneyLimit       //Limits of money available for algorithm
	algorithm       *entity.Algorithm          //Link to original object which algorithm based
	param           map[string]string          //Map of different algorithm configuration parameters (order expiration time etc)
	aChan           chan *stmodel.ActionReq    //Channel to send order requests to trader
	arChan          chan *stmodel.ActionResp   //Channel to receive responses from trader about action result
	stopCh          chan bool                  //Channel to stop algorithm when required
	buyPrice        map[string]decimal.Decimal //Cache of buy prices made previously (when sell goes after buy - it clears record) - to prevent selling cheaper than previous buy
	ordExp          time.Duration              //Expiration duration of posted orders - when expiration time passed and order not finished then it will be canceled
	commission      decimal.Decimal            //Commission on deals to take into account
	relDerivative   decimal.Decimal
	stopLossEnabled bool
	stopLossRel     decimal.Decimal //Relative price limit, when crossed - process market sell
	ctx             context.Context
	cancelF         context.CancelFunc
	instrAmount     map[string]int64 //Initial amount of instruments available

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
	RelDerivative   string = "relative_derivative"
	StopLoss        string = "stop_loss"
)

type AlgoData struct {
	statusMap   map[string]algoStatus //Algorithm status by every instrument (not block all algorithm with one instrument operation)
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
	statusMap := make(map[string]algoStatus)
	for _, figi := range a.figis {
		statusMap[figi] = process
	}
	aDat := AlgoData{ //Algorithm data storing as single thread state
		statusMap:   statusMap,
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
	if resp.Action.Status == entity.Success {
		iAmount := action.LotAmount
		if action.Direction == entity.Sell {
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
	aDat.statusMap[action.InstrFigi] = process
	a.updateState()
	return nil
}

func (a *AlgorithmImpl) processData(aDat *AlgoData, pDat *procData) {
	prevDiff, prevExists := aDat.prev[pDat.Figi]
	currDiff := pDat.SAV.Sub(pDat.LAV)
	relDer := pDat.DER.Div(pDat.SAV)
	a.logger.Debugf("Difference, current: %s, prev: %s, price: %s, derivative: %s, rel derivative: %s",
		currDiff, prevDiff, pDat.Price, pDat.DER, relDer)
	aDat.prev[pDat.Figi] = currDiff
	status, ok := aDat.statusMap[pDat.Figi]
	if !ok {
		aDat.statusMap[pDat.Figi] = process
	} else if status != process {
		a.logger.Debugf("Waiting in status: %d", status)
		return
	}
	buyPrice, ok := a.buyPrice[pDat.Figi]

	//If previous difference value not exists finish method
	if !prevExists {
		return
	}
	switch true {
	case pDat.DER.IsPositive() && ((prevDiff.IsNegative() && currDiff.IsPositive()) ||
		(prevDiff.IsPositive() && currDiff.IsPositive() && relDer.GreaterThan(a.relDerivative))):
		//Buy if exists previous difference by figi, short window derivative is positive AND
		//Go from negative to positive difference (short window crossing long) OR price growing now fast enough (by rel derivative setting)
		if ok {
			a.logger.Info("Previous buy operation not finished with price: ", buyPrice, "; waiting for sell operation...")
		} else {
			a.doBuy(aDat, pDat)
		}
	case pDat.DER.IsNegative() && (!ok || buyPrice.LessThan(pDat.Price)) &&
		((prevDiff.IsPositive() && currDiff.IsNegative()) || (currDiff.IsNegative() && prevDiff.IsNegative())):
		//Sell if exists previous difference by figi, short window derivative is negative, buy price not found or lower than current AND
		//Go from positive to negative difference (short window crossing long) OR price dropping
		if ok {
			buyPriceComm := buyPrice.Mul(decimal.NewFromInt(1).Add(a.commission.Mul(decimal.NewFromInt(2))))
			a.logger.Infof("Buy price found. Current price: %s, buy price: %s, buy with percents: %s", pDat.Price, buyPrice, buyPriceComm)
			if buyPriceComm.GreaterThanOrEqual(pDat.Price) {
				a.logger.Infof("Buy price %s is greater than current %s plus 2x commissions - not good enough, waiting better...",
					buyPriceComm, pDat.Price)
				break
			}
		}
		a.doSell(aDat, pDat, entity.Limited)
	case ok && pDat.DER.IsNegative() && a.stopLossEnabled && pDat.SAV.LessThanOrEqual(buyPrice.Mul(a.stopLossRel)):
		//If current price lower than stop loss - sell using market order
		a.logger.Infof("Stop loss reached; Current price: %s, stop loss price: %s", pDat.Price, buyPrice.Mul(a.stopLossRel))
		a.doSell(aDat, pDat, entity.Market)
	}
}

func (a *AlgorithmImpl) doBuy(aDat *AlgoData, pDat *procData) {
	action := entity.Action{
		AlgorithmID:    a.id,
		Direction:      entity.Buy,
		InstrFigi:      pDat.Figi,
		ReqPrice:       pDat.Price,
		ExpirationTime: time.Now().Add(a.ordExp),
		Status:         entity.Created,
		OrderType:      entity.Limited,
		RetrievedAt:    pDat.Time,
		AccountID:      a.accountId,
	}
	a.logger.Infof("Conditions for Buy, requesting action: %+v", action)
	a.aChan <- a.makeReq(&action)
	aDat.statusMap[pDat.Figi] = waitRes
}

func (a *AlgorithmImpl) doSell(aDat *AlgoData, pDat *procData, orderType entity.OrderType) {
	amount, iExists := aDat.instrAmount[pDat.Figi]
	if iExists && amount != 0 {
		action := entity.Action{
			AlgorithmID:    a.id,
			Direction:      entity.Sell,
			InstrFigi:      pDat.Figi,
			LotAmount:      amount,
			ReqPrice:       pDat.Price,
			ExpirationTime: time.Now().Add(a.ordExp),
			Status:         entity.Created,
			OrderType:      orderType,
			RetrievedAt:    pDat.Time,
			AccountID:      a.accountId,
		}
		a.logger.Infof("Conditions for Sell, requesting action: %+v", action)
		a.aChan <- a.makeReq(&action)
		aDat.statusMap[pDat.Figi] = waitRes
	}
}

//updateState updates algorithm context parameters from current state
func (a *AlgorithmImpl) updateState() {
	instruments := make([]*dto.InstrumentInfo, 0)
	for figi, amount := range a.instrAmount {
		info := dto.InstrumentInfo{
			Figi:   figi,
			Amount: amount,
		}
		if price, ok := a.buyPrice[figi]; ok {
			info.BuyPosPrice = price
		}
		instruments = append(instruments, &info)
	}
	res, err := json.Marshal(instruments)
	if err == nil {
		param, ok := a.algorithm.GetCtxParam(dto.InstrAmountField)
		if !ok {
			param = &entity.CtxParam{
				ID:          0,
				AlgorithmID: a.id,
				Key:         dto.InstrAmountField,
				Value:       string(res),
			}
			a.algorithm.CtxParams = append(a.algorithm.CtxParams, param)
		} else {
			param.Value = string(res)
		}
	}
}

func (a *AlgorithmImpl) makeReq(action *entity.Action) *stmodel.ActionReq {
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

func (a *AlgorithmImpl) Configure(ctx []*entity.CtxParam) error {
	if ctx == nil {
		a.logger.Infof("Algorihtm %d configuration not set, skipping", a.id)
		return nil
	}
	ctxParam := entity.ContextToMap(ctx)

	return configure(ctxParam, &algoState{
		InitAmount: a.instrAmount,
		BuyPrice:   a.buyPrice,
	}, a.logger)
}

func (a *AlgorithmImpl) GetParam() map[string]string {
	return a.param
}

func (a *AlgorithmImpl) GetAlgorithm() *entity.Algorithm {
	return a.algorithm
}

//NewProd constructs new algorithm using production data processor
func NewProd(algo *entity.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	proc, err := newDataProc(algo, infoSrv, logger)
	if err != nil {
		return nil, err
	}
	return newAvr(algo, logger, proc)
}

//NewSandbox constructs new algorithm using production data processor cause it the same for such algorithm
func NewSandbox(algo *entity.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	proc, err := newDataProc(algo, infoSrv, logger)
	if err != nil {
		return nil, err
	}
	return newAvr(algo, logger, proc)
}

//NewHist constructs new algorithm using history data processor
func NewHist(algo *entity.Algorithm, hRep repository.HistoryRepository, rootLogger *zap.SugaredLogger) (stmodel.Algorithm, error) {
	//New logger with Increased level to suppress history analysis not necessary logging
	logger := zap.New(rootLogger.Desugar().Core(), zap.IncreaseLevel(zap.WarnLevel)).Sugar()
	proc, err := newHistoryDataProc(algo, hRep, logger)
	if err != nil {
		return nil, err
	}
	return newAvr(algo, logger, proc)
}

//Main average algorithm constructor
func newAvr(algo *entity.Algorithm, logger *zap.SugaredLogger, proc DataProc) (stmodel.Algorithm, error) {
	//Turn params to map for convenience
	paramMap := entity.ParamsToMap(algo.Params)
	//Set order expiration time in seconds (when using limited requests), default 5 min
	ordExpInt := getOrDefaultInt(paramMap, OrderExpiration, 300)
	//Get stop loss parameter
	stopLossPercent, stopLossEnabled := getDecimal(paramMap, StopLoss)
	stopLossC := decimal.NewFromInt(1).Sub(stopLossPercent.Div(decimal.NewFromInt(100)))
	algorthm := &AlgorithmImpl{
		id:              algo.ID,
		isActive:        abool.NewBool(true),
		accountId:       algo.AccountId,
		dataProc:        proc,
		figis:           algo.Figis,
		limits:          algo.MoneyLimits,
		param:           paramMap,
		algorithm:       algo,
		buyPrice:        make(map[string]decimal.Decimal),
		stopCh:          make(chan bool),
		logger:          logger,
		relDerivative:   getOrDefaultDecimal(paramMap, RelDerivative, decimal.NewFromFloat(0.01)),
		ordExp:          time.Duration(ordExpInt) * time.Second,
		commission:      getOrDefaultDecimal(paramMap, Commission, decimal.NewFromFloat(0.04)).Div(decimal.NewFromInt(100)),
		stopLossRel:     stopLossC,
		stopLossEnabled: stopLossEnabled,
		instrAmount:     make(map[string]int64),
	}
	if err := algorthm.Configure(algo.CtxParams); err != nil {
		logger.Errorf("Failed configure algorithm %d with configuration %+v", algo.ID, algo.CtxParams)
		return nil, err
	}

	return algorthm, nil
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

func getDecimal(paramMap map[string]string, param string) (decimal.Decimal, bool) {
	res, ok := paramMap[param]
	if !ok {
		return decimal.Zero, false
	}
	resDec, err := decimal.NewFromString(res)
	if err != nil {
		return decimal.Zero, false
	}
	return resDec, true
}
