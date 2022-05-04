package trade

import "invest-robot/strategy/stmodel"

type Trader interface {
	AddSubscription(sub *stmodel.Subscription) error
	RemoveSubscription(id uint) error
}

func NewProdApiTrader() Trader {
	return &ProdApiTrader{}
}

func NewSandboxTrader() Trader {
	return &SandboxTrader{}
}
