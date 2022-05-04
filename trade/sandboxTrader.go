package trade

import (
	"invest-robot/errors"
	"invest-robot/strategy/stmodel"
)

type SandboxTrader struct {
}

func (at *SandboxTrader) AddSubscription(sub *stmodel.Subscription) error {
	return errors.NewNotImplemented()
}

func (at *SandboxTrader) RemoveSubscription(id uint) error {
	return errors.NewNotImplemented()
}
