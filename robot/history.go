package robot

import (
	"invest-robot/errors"
	investapi "invest-robot/tapigen"
	"time"
)

type HistoryAPI interface {
	LoadHistory(figis []string, startTime time.Time, endTime time.Time) error
}

type DefaultHistoryAPI struct {
	dsrv investapi.MarketDataServiceClient
}

func (h DefaultHistoryAPI) LoadHistory(figis []string, startTime time.Time, endTime time.Time) error {
	return errors.NewUnexpectedError("Not implemented yet!") //todo implement me!
}

func NewHistoryAPI() HistoryAPI {
	return DefaultHistoryAPI{}
}
