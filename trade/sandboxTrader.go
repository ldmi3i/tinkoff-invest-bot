package trade

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"invest-robot/collections"
	"invest-robot/domain"
	"invest-robot/dto/tapi"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy/stmodel"
	"invest-robot/trade/trmodel"
	"time"
)

type SandboxTrader struct {
	infoSrv   service.InfoSrv
	tradeSrv  service.TradeService
	actionRep repository.ActionRepository
	subs      collections.SyncMap[uint, *stmodel.Subscription]
	orders    collections.SyncMap[string, *domain.Action]

	algoCh chan *stmodel.ActionReq
	logger *zap.SugaredLogger
}

func (t *SandboxTrader) Go() {
	go t.checkOrdersBg()
	go t.actionProcBg()
}

func (t *SandboxTrader) AddSubscription(sub *stmodel.Subscription) error {
	t.logger.Infof("Add subscription for algo with id: %d", sub.AlgoID)
	t.subs.Put(sub.AlgoID, sub)

	go t.subBg(sub)
	return nil
}

//Background task to process subscriptions and redirect all of them to single stream
func (t *SandboxTrader) subBg(sub *stmodel.Subscription) {
	t.logger.Infof("Starting background processing for algo with id: %d", sub.AlgoID)
	//TODO Add stop channel and select
	for req := range sub.AChan {
		t.logger.Infof("Received algo request: %+v", req)
		t.algoCh <- req
	}
	t.logger.Infof("Stopping subscription for algo with id: %d", sub.AlgoID)
}

func (t *SandboxTrader) checkOrdersBg() {
	t.logger.Info("Starting background checking orders...")
	for {
		sl := t.orders.GetSlice()
		t.logger.Debug("Check orders, len ", len(sl))
		for _, entry := range sl {
			action := entry.Value
			req := tapi.GetOrderStateRequest{
				AccountId: action.AccountID,
				OrderId:   entry.Key,
			}
			state, err := t.infoSrv.GetOrderState(&req)
			if err != nil {
				t.logger.Errorf("Error checking order state %+v: %s", req, err)
				continue
			}
			t.logger.Info("Check order with id ", entry.Key, " status ", state.ExecStatus)
			switch state.ExecStatus {
			case tapi.EXECUTION_REPORT_STATUS_FILL:
				action.Status = domain.SUCCESS
				action.Info = "Order successfully completed"
				action.Amount = state.TotalPrice.Value
				action.Currency = state.TotalPrice.Currency
				action.PositionPrice = state.AvrPrice.Value
				err = t.actionRep.Save(action)
				if err != nil {
					t.logger.Errorf("Error while updating action %+v: %s", action, err)
				}
				t.orders.Delete(entry.Key)
				t.logger.Info("Order with id ", entry.Key, " completed")
				sub, ok := t.subs.Get(action.AlgorithmID)
				if !ok {
					t.logger.Warn("Subscription by id ", action.ID, " not found")
					continue
				}
				sub.RChan <- &stmodel.ActionResp{Action: action}
			case tapi.EXECUTION_REPORT_STATUS_REJECTED:
				action.Status = domain.FAILED
				action.Info = "Order was rejected"
				err = t.actionRep.Save(action)
				if err != nil {
					t.logger.Errorf("Error while updating action %+v : %s", action, err)
				}
				t.orders.Delete(entry.Key)
				t.logger.Infof("Order with id %s rejected", entry.Key)
				sub, ok := t.subs.Get(action.ID)
				if !ok {
					t.logger.Warnf("Subscription by id %d not found", action.ID)
					continue
				}
				sub.RChan <- &stmodel.ActionResp{Action: action}
			case tapi.EXECUTION_REPORT_STATUS_CANCELLED:
				action.Status = domain.FAILED
				action.Info = "Order was canceled"
				err = t.actionRep.Save(action)
				if err != nil {
					t.logger.Errorf("Error while updating action %+v: %s", action, err)
				}
				t.orders.Delete(entry.Key)
				t.logger.Infof("Order with id %s rejected", entry.Key)
				sub, ok := t.subs.Get(action.ID)
				if !ok {
					t.logger.Warnf("Subscription by id %d not found", action.ID)
					continue
				}
				sub.RChan <- &stmodel.ActionResp{Action: action}
			}
		}
		time.Sleep(30 * time.Second)
	}
}

//Background task to process actions from algorithm
func (t *SandboxTrader) actionProcBg() {
	defer func() {
		if pnc := recover(); pnc != nil {
			t.logger.Error("Action bg task recovered ", pnc)
		}
	}()
	t.logger.Info("Background action listening started...")
	for req := range t.algoCh {
		t.logger.Info("Received action request: ", req)
		action := req.Action
		subscription, ok := t.subs.Get(action.AlgorithmID)
		if !ok {
			t.logger.Warnf("Error - subscription related to action not found, algo id: %d", action.AlgorithmID)
			continue
		}
		opInfo, ok := t.preprocessAction(req, subscription)
		if !ok {
			continue
		}
		if action.Direction == domain.BUY {
			t.procBuy(opInfo, action, subscription)
		} else {
			t.procSell(opInfo, action, subscription)
		}
	}
	t.logger.Info("Background action listening finished...")
}

//Validate parameters and populate info context for ordering
//Returns information and flag: true if validation succeed, false if validation failed and no order may be placed
func (t *SandboxTrader) preprocessAction(req *stmodel.ActionReq, subscription *stmodel.Subscription) (*trmodel.OpInfo, bool) {
	action := req.Action
	err := t.actionRep.Save(action)
	if err != nil {
		t.logger.Error("Error while saving action. Canceling operation... ", err)
		t.setActionStatus(action, domain.FAILED, "Error while saving action")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	//Retrieving instrument for order
	instrInfo, err := t.infoSrv.GetInstrumentInfoByFigi(action.InstrFigi)
	if err != nil {
		t.logger.Error("Error while requesting instrument info. Canceling operation, updating status...", err)
		t.setActionStatus(action, domain.FAILED, "Error getting instrument info")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	action.Currency = instrInfo.Currency
	//Check api available flag
	if !instrInfo.ApiTradeAvailableFlag {
		t.logger.Errorf("Instrument with figi %s not available for trading through API", action.InstrFigi)
		t.setActionStatus(action, domain.FAILED, "Instrument operating through API not available")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	//Check is specific operation available for instrument
	if (!instrInfo.SellAvailableFlag && action.Direction == domain.SELL) ||
		(!instrInfo.BuyAvailableFlag && action.Direction == domain.BUY) {
		t.logger.Errorf("Operation by instrument not available...")
		t.setActionStatus(action, domain.FAILED, "Operation by instrument not available")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	//Check is trade session has ok status
	if !instrInfo.IsTradingAvailable() {
		t.logger.Warn("Exchange trading status has incorrect status.", instrInfo.TradingStatus)
		t.setActionStatus(action, domain.FAILED, fmt.Sprintf("Exchange has incorrect status %d", instrInfo.TradingStatus))
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	opInfo := trmodel.OpInfo{
		Currency: action.Currency, Lim: req.GetCurrLimit(action.Currency), PosNum: instrInfo.Lot}
	if opInfo.Lim.IsZero() {
		t.logger.Warnf("Limit for currency %s not set, discarding order", action.Currency)
		t.setActionStatus(action, domain.FAILED, "Limit by requested currency not set")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	//Retrieve last single lot price
	prices, err := t.infoSrv.GetLastPrices([]string{action.InstrFigi})
	if err != nil || prices.GetByFigi(action.InstrFigi) == nil {
		t.logger.Error("Error retrieving last prices by ", instrInfo.TradingStatus)
		t.setActionStatus(action, domain.FAILED, "Error getting price by figi")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	opInfo.PosPrice = prices.GetByFigi(action.InstrFigi).Price
	t.logger.Debugf("Preprocess for action %d finished", action.ID)
	return &opInfo, true
}

//Process buy order
func (t *SandboxTrader) procBuy(opInfo *trmodel.OpInfo, action *domain.Action, sub *stmodel.Subscription) {
	t.logger.Debug("Starting buy for action ", action.ID)
	//Calculating price for single buy operation multiple to instrument weight
	lotPrice := decimal.NewFromInt(opInfo.PosNum).Mul(opInfo.PosPrice)
	//Check is minimum instrument price exceed the limit
	if lotPrice.GreaterThan(opInfo.Lim) {
		t.logger.Warnf("Limit lower than minimal buy price, figi %s; limit: %s; lot price: %s; one price: %s",
			action.InstrFigi, opInfo.Lim, opInfo.PosPrice, lotPrice)
		t.setActionStatus(action, domain.FAILED, "Price of one buy exceeds limit")
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	//Calculate number of weighted instruments possible to buy for existing limit
	operNum := opInfo.Lim.Div(lotPrice).Floor()
	//Calculate required amount of money for this order
	moneyAmount := operNum.Mul(lotPrice)
	//Calculate instrument amount to buy
	lotAmount := operNum.IntPart()
	//Get real available money amount using GetPositions request
	posReq := tapi.PositionsRequest{AccountId: action.AccountID}
	positions, err := t.infoSrv.GetPositions(&posReq)
	if err != nil {
		t.logger.Error("Error getting positions ", err)
		t.setActionStatus(action, domain.FAILED, "Error while getting positions")
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	moneyAvail := positions.GetMoney(action.Currency)
	//Check is account has available amount of money for operation
	if moneyAvail == nil || moneyAvail.Value.LessThan(moneyAmount) {
		t.logger.Warnf("Not enough money for figi %s;  lot price: %s; lot num: %d; required money: %s; available money: %s",
			action.InstrFigi, opInfo.PosPrice, opInfo.PosNum, moneyAmount, moneyAvail)
		t.setActionStatus(action, domain.FAILED, fmt.Sprintf("No money for operation"))
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	//Prepare request (currently only Market order type) and post order
	orderId := uuid.New().String()
	req := tapi.PostOrderRequest{
		Figi:      action.InstrFigi,
		PosNum:    lotAmount,
		Direction: tapi.ORDER_DIRECTION_BUY,
		AccountId: action.AccountID,
		OrderType: tapi.ORDER_TYPE_MARKET,
		OrderId:   orderId,
	}
	action.OrderId = orderId
	order, err := t.tradeSrv.PostOrder(&req)
	if err != nil {
		t.logger.Errorf("Error posting buy order %s: %s", orderId, err)
		t.setActionStatus(action, domain.FAILED, "Error while posting buy order")
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	t.logger.Infof("Posted buy order: %+v", order)
	//Set amounts of money to action and update action status in db
	action.Amount = moneyAmount
	action.LotAmount = lotAmount
	t.orders.Put(order.OrderId, action)
	t.setActionStatus(action, domain.POSTED, "Action posted successfully")
}

//Process sell order (currently there's no limits for a sell operation)
func (t *SandboxTrader) procSell(opInfo *trmodel.OpInfo, action *domain.Action, sub *stmodel.Subscription) {
	//Check is requested instrument amount for a sell not zero
	if action.LotAmount == 0 {
		t.logger.Warn("LotAmount is 0 - nothing to sell")
		t.setActionStatus(action, domain.FAILED, "No instrument to sell found")
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	//Posting market sell request
	orderId := uuid.New().String()
	req := tapi.PostOrderRequest{
		Figi:      action.InstrFigi,
		PosNum:    action.LotAmount,
		Direction: tapi.ORDER_DIRECTION_SELL,
		AccountId: action.AccountID,
		OrderType: tapi.ORDER_TYPE_MARKET,
		OrderId:   orderId,
	}
	action.OrderId = orderId
	order, err := t.tradeSrv.PostOrder(&req)
	if err != nil {
		t.logger.Errorf("Error posting sell order %s: %s", orderId, err)
		t.setActionStatus(action, domain.FAILED, "Error while posting buy order")
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	//Populating orders map to further state monitoring and responding
	t.orders.Put(order.OrderId, action)
	t.logger.Info("Posted sell order ", order)
	t.setActionStatus(action, domain.POSTED, "Sell order successfully posted")
}

//Set status and info message to action and updates it in db
func (t *SandboxTrader) setActionStatus(action *domain.Action, status domain.ActionStatus, msg string) {
	action.Status = status
	if err := t.actionRep.UpdateStatusWithMsg(action.ID, action.Status, msg); err != nil {
		t.logger.Error("Error while updating status, skipping update...", err)
	}
}

// RemoveSubscription removes algorithm subscription and stop monitoring
func (t *SandboxTrader) RemoveSubscription(id uint) error {
	t.logger.Infof("Remove subscription for algo with id: %d", id)
	sub, ok := t.subs.Get(id)
	if !ok {
		return nil
	}
	close(sub.RChan)
	t.subs.Delete(id)

	return nil
}

func NewSandboxTrader(infoSrv service.InfoSrv, tradeSrv service.TradeService, actionRep repository.ActionRepository, logger *zap.SugaredLogger) Trader {
	return &SandboxTrader{
		infoSrv:   infoSrv,
		tradeSrv:  tradeSrv,
		actionRep: actionRep,
		subs:      collections.NewSyncMap[uint, *stmodel.Subscription](),
		orders:    collections.NewSyncMap[string, *domain.Action](),
		algoCh:    make(chan *stmodel.ActionReq, 1),
		logger:    logger,
	}
}
