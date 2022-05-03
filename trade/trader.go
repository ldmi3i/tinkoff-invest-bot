package trade

import (
	"invest-robot/dto"
	"invest-robot/repository"
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

func NewMockTrader(hRep repository.HistoryRepository) MockTrader {
	return MockTrader{statCh: make(chan dto.HistStatResponse), hRep: hRep}
}
