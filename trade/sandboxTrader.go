package trade

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"invest-robot/collections"
	"invest-robot/domain"
	"invest-robot/dto/tapi"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy/stmodel"
	"invest-robot/trade/trmodel"
	"log"
	"time"
)

type SandboxTrader struct {
	infoSrv   service.InfoSrv
	tradeSrv  service.TradeService
	actionRep repository.ActionRepository
	subs      collections.SyncMap[uint, *stmodel.Subscription]
	orders    collections.SyncMap[string, *domain.Action]

	algoCh chan *stmodel.ActionReq
}

func (t *SandboxTrader) Go() {
	go t.checkOrdersBg()
	go t.actionProcBg()
}

func (t *SandboxTrader) AddSubscription(sub *stmodel.Subscription) error {
	log.Printf("Add subscription for algo with id: %d", sub.AlgoID)
	t.subs.Put(sub.AlgoID, sub)

	go t.subBg(sub)
	return nil
}

//Background task to process subscriptions and redirect all of them to single stream
func (t *SandboxTrader) subBg(sub *stmodel.Subscription) {
	log.Printf("Starting background processing for algo with id: %d", sub.AlgoID)
	//TODO Add stop channel and select
	for req := range sub.AChan {
		log.Printf("Received algo request: %+v", req)
		t.algoCh <- req
	}
	log.Printf("Stopping subscription for algo with id: %d", sub.AlgoID)
}

func (t *SandboxTrader) checkOrdersBg() {
	log.Println("Starting background checking orders...")
	for {
		sl := t.orders.GetSlice()
		log.Println("Check orders, len", len(sl))
		for _, entry := range sl {
			action := entry.Value
			req := tapi.GetOrderStateRequest{
				AccountId: action.AccountID,
				OrderId:   entry.Key,
			}
			state, err := t.infoSrv.GetOrderState(&req)
			if err != nil {
				log.Println("Error checking order state", req, err)
				continue
			}
			log.Println("Check order with id", entry.Key, "status", state.ExecStatus)
			switch state.ExecStatus {
			case tapi.EXECUTION_REPORT_STATUS_FILL:
				action.Status = domain.SUCCESS
				action.Info = "Order successfully completed"
				action.Amount = state.TotalPrice.Value
				action.Currency = state.TotalPrice.Currency
				err = t.actionRep.Save(action)
				if err != nil {
					log.Println("Error while updating action", action, err)
				}
				t.orders.Delete(entry.Key)
				log.Println("Order with id", entry.Key, "completed")
				sub, ok := t.subs.Get(action.AlgorithmID)
				if !ok {
					log.Println("Subscription by id", action.ID, "not found")
					continue
				}
				sub.RChan <- &stmodel.ActionResp{Action: action}
			case tapi.EXECUTION_REPORT_STATUS_REJECTED:
				action.Status = domain.FAILED
				action.Info = "Order was rejected"
				err = t.actionRep.Save(action)
				if err != nil {
					log.Println("Error while updating action", action, err)
				}
				t.orders.Delete(entry.Key)
				log.Println("Order with id", entry.Key, "rejected")
				sub, ok := t.subs.Get(action.ID)
				if !ok {
					log.Println("Subscription by id", action.ID, "not found")
					continue
				}
				sub.RChan <- &stmodel.ActionResp{Action: action}
			case tapi.EXECUTION_REPORT_STATUS_CANCELLED:
				action.Status = domain.FAILED
				action.Info = "Order was canceled"
				err = t.actionRep.Save(action)
				if err != nil {
					log.Println("Error while updating action", action, err)
				}
				t.orders.Delete(entry.Key)
				log.Println("Order with id", entry.Key, "rejected")
				sub, ok := t.subs.Get(action.ID)
				if !ok {
					log.Println("Subscription by id", action.ID, "not found")
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
			log.Println("Action bg task recovered ", pnc)
		}
	}()
	log.Println("Background action listening started...")
	for req := range t.algoCh {
		log.Println("Received action request:", req)
		action := req.Action
		subscription, ok := t.subs.Get(action.AlgorithmID)
		if !ok {
			log.Printf("Error - subscription related to action not found, algo id: %d", action.AlgorithmID)
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
	log.Println("Background action listening finished...")
}

//Validate parameters and populate info context for ordering
//Returns information and flag: true if validation succeed, false if validation failed and no order may be placed
func (t *SandboxTrader) preprocessAction(req *stmodel.ActionReq, subscription *stmodel.Subscription) (*trmodel.OpInfo, bool) {
	action := req.Action
	err := t.actionRep.Save(action)
	if err != nil {
		log.Println("Error while saving action. Canceling operation...", err)
		t.setActionStatus(action, domain.FAILED, "Error while saving action")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	//Retrieving instrument for order
	instrInfo, err := t.infoSrv.GetInstrumentInfoByFigi(action.InstrFigi)
	if err != nil {
		log.Println("Error while requesting instrument info. Canceling operation, updating status...", err)
		t.setActionStatus(action, domain.FAILED, "Error getting instrument info")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	action.Currency = instrInfo.Currency
	//Check api available flag
	if !instrInfo.ApiTradeAvailableFlag {
		log.Printf("ERROR Instrument with figi %s not available for trading through API", action.InstrFigi)
		t.setActionStatus(action, domain.FAILED, "Instrument operating through API not available")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	//Check is specific operation available for instrument
	if (!instrInfo.SellAvailableFlag && action.Direction == domain.SELL) ||
		(!instrInfo.BuyAvailableFlag && action.Direction == domain.BUY) {
		log.Printf("ERROR Operation by instrument not available...")
		t.setActionStatus(action, domain.FAILED, "Operation by instrument not available")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	//Check is trade session has ok status
	if !instrInfo.IsTradingAvailable() {
		log.Println("Exchange trading status has incorrect status.", instrInfo.TradingStatus)
		t.setActionStatus(action, domain.FAILED, fmt.Sprintf("Exchange has incorrect status %d", instrInfo.TradingStatus))
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	opInfo := trmodel.OpInfo{
		Currency: action.Currency, Lim: req.GetCurrLimit(action.Currency), PosNum: instrInfo.Lot}
	if opInfo.Lim.IsZero() {
		log.Println("Limit for currency", action.Currency, "not set, discarding order")
		t.setActionStatus(action, domain.FAILED, "Limit by requested currency not set")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	//Retrieve last single lot price
	prices, err := t.infoSrv.GetLastPrices([]string{action.InstrFigi})
	if err != nil || prices.GetByFigi(action.InstrFigi) == nil {
		log.Println("Error retrieving last prices by ", instrInfo.TradingStatus)
		t.setActionStatus(action, domain.FAILED, "Error getting price by figi")
		subscription.RChan <- &stmodel.ActionResp{Action: action}
		return nil, false
	}
	opInfo.LotPrice = prices.GetByFigi(action.InstrFigi).Price
	log.Println("Preprocess for action", action.ID, "finished")
	return &opInfo, true
}

//Process buy order
func (t *SandboxTrader) procBuy(opInfo *trmodel.OpInfo, action *domain.Action, sub *stmodel.Subscription) {
	log.Println("Starting buy for action", action.ID)
	//Calculating price for single buy operation multiple to instrument weight
	lotPrice := decimal.NewFromInt(opInfo.PosNum).Mul(opInfo.LotPrice)
	//Check is minimum instrument price exceed the limit
	if lotPrice.GreaterThan(opInfo.Lim) {
		log.Printf("Limit lower than minimal buy price, figi %s; limit: %s; lot price: %s; one price: %s",
			action.InstrFigi, opInfo.Lim, opInfo.LotPrice, lotPrice)
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
		log.Println("Error getting positions", err)
		t.setActionStatus(action, domain.FAILED, "Error while getting positions")
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	moneyAvail := positions.GetMoney(action.Currency)
	//Check is account has available amount of money for operation
	if moneyAvail == nil || moneyAvail.Value.LessThan(moneyAmount) {
		log.Printf("Not enough money for figi %s;  lot price: %s; lot num: %d; required money: %s; available money: %s",
			action.InstrFigi, opInfo.LotPrice, opInfo.PosNum, moneyAmount, moneyAvail)
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
		log.Println("Error posting buy order", orderId, err)
		t.setActionStatus(action, domain.FAILED, "Error while posting buy order")
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	log.Printf("Posted buy order: %+v", order)
	//Set amounts of money to action and update action status in db
	action.Amount = moneyAmount
	action.InstrAmount = lotAmount
	t.orders.Put(order.OrderId, action)
	t.setActionStatus(action, domain.POSTED, "Action posted successfully")
}

//Process sell order (currently there's no limits for a sell operation)
func (t *SandboxTrader) procSell(opInfo *trmodel.OpInfo, action *domain.Action, sub *stmodel.Subscription) {
	//Check is requested instrument amount for a sell not zero
	if action.InstrAmount == 0 {
		log.Println("InstrAmount is 0 - nothing to sell")
		t.setActionStatus(action, domain.FAILED, "No instrument to sell found")
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	//Check is requested instrument amount not lower than instrument lot weight
	if action.InstrAmount < opInfo.PosNum {
		log.Printf("Not enough lots for one operation; requested: %d; lot num: %d", action.InstrAmount, opInfo.PosNum)
		t.setActionStatus(action, domain.FAILED, "Not enough instrument for sell")
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	//Posting market sell request
	orderId := uuid.New().String()
	req := tapi.PostOrderRequest{
		Figi:      action.InstrFigi,
		PosNum:    action.InstrAmount,
		Direction: tapi.ORDER_DIRECTION_SELL,
		AccountId: action.AccountID,
		OrderType: tapi.ORDER_TYPE_MARKET,
		OrderId:   orderId,
	}
	action.OrderId = orderId
	order, err := t.tradeSrv.PostOrder(&req)
	if err != nil {
		log.Println("Error posting sell order", orderId)
		t.setActionStatus(action, domain.FAILED, "Error while posting buy order")
		sub.RChan <- &stmodel.ActionResp{Action: action}
		return
	}
	//Populating orders map to further state monitoring and responding
	t.orders.Put(order.OrderId, action)
	log.Println("Posted sell order", order)
	t.setActionStatus(action, domain.POSTED, "Sell order successfully posted")
}

//Set status and info message to action and updates it in db
func (t *SandboxTrader) setActionStatus(action *domain.Action, status domain.ActionStatus, msg string) {
	action.Status = status
	if err := t.actionRep.UpdateStatusWithMsg(action.ID, action.Status, msg); err != nil {
		log.Println("Error while updating status, skipping update...", err)
	}
}

// RemoveSubscription removes algorithm subscription and stop monitoring
func (t *SandboxTrader) RemoveSubscription(id uint) error {
	log.Printf("Remove subscription for algo with id: %d", id)
	sub, ok := t.subs.Get(id)
	if !ok {
		return nil
	}
	close(sub.RChan)
	t.subs.Delete(id)

	return nil
}

func NewSandboxTrader(infoSrv service.InfoSrv, tradeSrv service.TradeService, actionRep repository.ActionRepository) Trader {
	return &SandboxTrader{
		infoSrv:   infoSrv,
		tradeSrv:  tradeSrv,
		actionRep: actionRep,
		subs:      collections.NewSyncMap[uint, *stmodel.Subscription](),
		orders:    collections.NewSyncMap[string, *domain.Action](),
		algoCh:    make(chan *stmodel.ActionReq, 1),
	}
}
