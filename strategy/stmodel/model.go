package stmodel

import (
	"github.com/shopspring/decimal"
	"invest-robot/domain"
)

//Algorithm is a general interface of any trading logic
//It supposed to run in the background and processing data, retrieved from data processor through channel
//Exchange with Trader made through channels from Subscription object.
type Algorithm interface {
	//Configure is to configure Algorithm after restoring it state and data from db etc.
	Configure(ctx []domain.CtxParam) error
	Subscribe() (*Subscription, error)
	IsActive() bool
	GetId() uint
	Go() error
	Stop() error
}

type ActionResp struct {
	Action *domain.Action
}

type ActionReq struct {
	Action *domain.Action
	Limits []*domain.MoneyLimit
}

func (req ActionReq) GetCurrLimit(currency string) decimal.Decimal {
	for _, limit := range req.Limits {
		if currency == limit.Currency {
			return limit.Amount
		}
	}
	return decimal.Zero
}

type Subscription struct {
	AlgoID uint
	AChan  <-chan *ActionReq
	RChan  chan<- *ActionResp
}

type TraderData struct {
	Figi  string
	Price decimal.Decimal
}
