package tinapi

import (
	"invest-robot/domain"
	investapi "invest-robot/tapigen"
	"time"
)

type TinApi interface {
	GetOrderBook() (*investapi.GetOrderBookResponse, error)
	GetHistory(figis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time) ([]domain.History, error)
	GetDataStream() (*investapi.MarketDataStreamService_MarketDataStreamClient, error)
}
