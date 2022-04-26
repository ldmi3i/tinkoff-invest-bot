package tinapi

import (
	investapi "invest-robot/tapigen"
	"time"
)

type TinApi interface {
	GetOrderBook() (*investapi.GetOrderBookResponse, error)
	GetHistory(figis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time) ([]*investapi.GetCandlesResponse, error)
}
