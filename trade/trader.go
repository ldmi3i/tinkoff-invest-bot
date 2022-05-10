package trade

import "invest-robot/strategy/stmodel"

type Trader interface {
	AddSubscription(sub *stmodel.Subscription) error
	RemoveSubscription(id uint) error
	Go()
}
