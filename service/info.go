package service

import (
	"invest-robot/domain"
	investapi "invest-robot/tapigen"
	"invest-robot/tinapi"
	"time"
)

type InfoSrv interface {
	GetOrderBook() (*investapi.GetOrderBookResponse, error)
	GetHistory(finis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time) ([]domain.History, error)
}

type DefaultInfoSrv struct {
	tapi tinapi.TinApi
}

func NewInfoService(t tinapi.TinApi) InfoSrv {
	return DefaultInfoSrv{t}
}

func (i DefaultInfoSrv) GetOrderBook() (*investapi.GetOrderBookResponse, error) {
	return i.tapi.GetOrderBook()
}

func (i DefaultInfoSrv) GetHistory(finis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time) ([]domain.History, error) {
	respArr, err := i.tapi.GetHistory(finis, ivl, startDate, endDate)
	if err != nil {
		return nil, err
	}
	hist := make([]domain.History, 0)
	for _, resp := range respArr {
		for _, cndl := range resp.GetCandles() {
			histRec := domain.FromHistoricCandle(cndl)
			hist = append(hist, histRec)
		}
	}
	return hist, nil
}
