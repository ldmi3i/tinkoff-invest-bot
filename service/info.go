package service

import (
	"invest-robot/domain"
	"invest-robot/errors"
	investapi "invest-robot/tapigen"
	"invest-robot/tinapi"
	"time"
)

type InfoSrv interface {
	GetOrderBook() (*investapi.GetOrderBookResponse, error)
	//GetHistorySorted return sorted by time history in time interval
	GetHistorySorted(finis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time) ([]domain.History, error)
	//GetDataStream returns bidirectional data stream client
	GetDataStream() (investapi.MarketDataStreamService_MarketDataStreamClient, error)
	//GetAllShares return all shares, available for operating through API
	GetAllShares() (*investapi.SharesResponse, error)
}

type DefaultInfoSrv struct {
	tapi tinapi.TinApi
}

func NewInfoService(t tinapi.TinApi) InfoSrv {
	return &DefaultInfoSrv{t}
}

func (i *DefaultInfoSrv) GetOrderBook() (*investapi.GetOrderBookResponse, error) {
	return i.tapi.GetOrderBook()
}

func (i *DefaultInfoSrv) GetHistorySorted(figis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time) ([]domain.History, error) {
	next, err := nextTime(ivl, startTime)
	if err != nil {
		return nil, err
	}
	var hist []domain.History
	if next.After(endTime) {
		hist, err = i.tapi.GetHistorySorted(figis, ivl, startTime, endTime)
		if err != nil {
			return nil, err
		}
	} else {
		if hist, err = i.tapi.GetHistorySorted(figis, ivl, startTime, *next); err != nil {
			return nil, err
		}
		curr := next
		if next, err = nextTime(ivl, *next); err != nil {
			return nil, err
		}
		for next.Before(endTime) {
			row, err := i.tapi.GetHistorySorted(figis, ivl, *curr, *next)
			if err != nil {
				return nil, err
			}
			hist = append(hist, row...)
			curr = next
			if next, err = nextTime(ivl, *next); err != nil {
				return nil, err
			}
		}
		row, err := i.tapi.GetHistorySorted(figis, ivl, *curr, endTime)
		if err != nil {
			return nil, err
		}
		hist = append(hist, row...)
	}
	return hist, nil
}

func nextTime(ivl investapi.CandleInterval, startTime time.Time) (*time.Time, error) {
	switch ivl {
	case investapi.CandleInterval_CANDLE_INTERVAL_1_MIN,
		investapi.CandleInterval_CANDLE_INTERVAL_5_MIN,
		investapi.CandleInterval_CANDLE_INTERVAL_15_MIN:
		next := startTime.AddDate(0, 0, 1)
		return &next, nil
	case investapi.CandleInterval_CANDLE_INTERVAL_HOUR:
		next := startTime.AddDate(0, 0, 7)
		return &next, nil
	case investapi.CandleInterval_CANDLE_INTERVAL_DAY:
		next := startTime.AddDate(1, 0, 0)
		return &next, nil
	default:
		return nil, errors.NewUnexpectedError("Unexpected candle interval: " + ivl.String())
	}
}

func (i *DefaultInfoSrv) GetDataStream() (investapi.MarketDataStreamService_MarketDataStreamClient, error) {
	return i.tapi.GetDataStream()
}

func (i *DefaultInfoSrv) GetAllShares() (*investapi.SharesResponse, error) {
	return i.tapi.GetAllShares()
}
