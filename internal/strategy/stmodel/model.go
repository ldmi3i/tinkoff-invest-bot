package stmodel

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/shopspring/decimal"
)

//Algorithm is a general interface of any trading logic
//It supposed to run in the background and processing data, retrieved from data processor through channel
//Exchange with Trader made through channels from Subscription object.
//Algorithm keeps communication with Trader through channels using Action domain model
//Action domain model keeps all required information about single order request
//Action for sending from algorithm must be populated only partially
//some parameters populated and persisted by Trader (see domain.Action documentation)
type Algorithm interface {
	//Configure is to configure Algorithm after restoring it state and data from db etc. (Currently not implemented)
	Configure(ctx []*entity.CtxParam) error
	//Subscribe to algorithm and retrieve subscription to interact with algorithm
	Subscribe() (*Subscription, error)
	//IsActive return true if algorithm running and false if it was stopped
	IsActive() bool
	//GetParam returns algorithm parameters as map
	GetParam() map[string]string
	//GetAlgorithm returns algorithm data which used as base
	GetAlgorithm() *entity.Algorithm
	//Go starts algorithm running in background
	Go(ctx context.Context) error
	//Stop running algorithm
	Stop() error
}

//ParamSplitter is a common interface
//for splitting algorithm params into range of param set for analysis purposes
//Each algorithm must have itself parameter set, delimiter and splitting logic
//This interface must be implemented for the algorithm and added to factory to
//provide possibility to use /analyze/range request on algorithm
type ParamSplitter interface {
	ParseAndSplit(param map[string]string) ([]map[string]string, error)
}

//ActionResp represents common trader response model
type ActionResp struct {
	Action *entity.Action
}

//ActionReq represents common algorithm request model
type ActionReq struct {
	Action *entity.Action
	Limits []*entity.MoneyLimit
}

//GetCurrLimit simplifies retrieving limit by required currency from limit slice
func (req ActionReq) GetCurrLimit(currency string) decimal.Decimal {
	for _, limit := range req.Limits {
		if currency == limit.Currency {
			return limit.Amount
		}
	}
	return decimal.Zero
}

//Subscription represents common subscription object representing Algorithm/trade.Trader interaction mechanisms
type Subscription struct {
	AlgoID uint               //Subscribing algorithm identity
	AChan  <-chan *ActionReq  //Algorithm -> trade.Trader channel - to create order requests
	RChan  chan<- *ActionResp //trade.Trader -> Algorithm channel - to retrieve order result responses
}
