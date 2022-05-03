package trade

import (
	"invest-robot/strategy/model"
)

type Trader interface {
	AddSubscription(sub *model.Subscription) error
	RemoveSubscription(id uint) error
}

func NewProdApiTrader() Trader {
	return &ProdApiTrader{}
}

func NewSandboxTrader() Trader {
	return &SandboxTrader{}
}
