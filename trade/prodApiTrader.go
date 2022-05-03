package trade

import (
	"invest-robot/errors"
	"invest-robot/strategy/model"
)

type ProdApiTrader struct {
}

func (at *ProdApiTrader) AddSubscription(sub *model.Subscription) error {
	return errors.NewNotImplemented()
}

func (at *ProdApiTrader) RemoveSubscription(id uint) error {
	return errors.NewNotImplemented()
}
