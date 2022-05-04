package trade

import (
	"invest-robot/errors"
	"invest-robot/strategy/stmodel"
)

type ProdApiTrader struct {
}

func (at *ProdApiTrader) AddSubscription(sub *stmodel.Subscription) error {
	return errors.NewNotImplemented()
}

func (at *ProdApiTrader) RemoveSubscription(id uint) error {
	return errors.NewNotImplemented()
}
