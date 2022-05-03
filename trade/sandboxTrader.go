package trade

import (
	"invest-robot/errors"
	"invest-robot/strategy/model"
)

type SandboxTrader struct {
}

func (at *SandboxTrader) AddSubscription(sub *model.Subscription) error {
	return errors.NewNotImplemented()
}

func (at *SandboxTrader) RemoveSubscription(id uint) error {
	return errors.NewNotImplemented()
}
