package service

import (
	"context"
	"invest-robot/domain"
	"invest-robot/dto/dtotapi"
	"invest-robot/errors"
	investapi "invest-robot/tapigen"
	"invest-robot/tinapi"
	"time"
)

type InfoSrv interface {
	//GetHistorySorted return sorted by time history in time interval
	GetHistorySorted(finis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time, ctx context.Context) ([]domain.History, error)
	//GetDataStream returns bidirectional data stream client
	GetDataStream(ctx context.Context) (investapi.MarketDataStreamService_MarketDataStreamClient, error)
	//GetAllShares return all shares, available for operating through API
	GetAllShares(ctx context.Context) (*investapi.SharesResponse, error) //TODO change response to DTO!
	GetInstrumentInfoByFigi(figi string, ctx context.Context) (*dtotapi.InstrumentResponse, error)
	GetOrderState(req *dtotapi.GetOrderStateRequest, ctx context.Context) (*dtotapi.GetOrderStateResponse, error)

	GetLastPrices(figis []string, ctx context.Context) (*dtotapi.LastPricesResponse, error)
	GetPositions(req *dtotapi.PositionsRequest, ctx context.Context) (*dtotapi.PositionsResponse, error)
}

type BaseInfoSrv struct {
	tapi tinapi.Api
}

func newBaseSrv(t tinapi.Api) *BaseInfoSrv {
	return &BaseInfoSrv{tapi: t}
}

func (i *BaseInfoSrv) GetHistorySorted(figis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time, ctx context.Context) ([]domain.History, error) {
	next, err := nextTime(ivl, startTime)
	if err != nil {
		return nil, err
	}
	var hist []domain.History
	if next.After(endTime) {
		hist, err = i.tapi.GetHistorySorted(figis, ivl, startTime, endTime, ctx)
		if err != nil {
			return nil, err
		}
	} else {
		if hist, err = i.tapi.GetHistorySorted(figis, ivl, startTime, *next, ctx); err != nil {
			return nil, err
		}
		curr := next
		if next, err = nextTime(ivl, *next); err != nil {
			return nil, err
		}
		for next.Before(endTime) {
			row, err := i.tapi.GetHistorySorted(figis, ivl, *curr, *next, ctx)
			if err != nil {
				return nil, err
			}
			hist = append(hist, row...)
			curr = next
			if next, err = nextTime(ivl, *next); err != nil {
				return nil, err
			}
		}
		row, err := i.tapi.GetHistorySorted(figis, ivl, *curr, endTime, ctx)
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

func (i *BaseInfoSrv) GetDataStream(ctx context.Context) (investapi.MarketDataStreamService_MarketDataStreamClient, error) {
	return i.tapi.MarketDataStream(ctx)
}

func (i *BaseInfoSrv) GetAllShares(ctx context.Context) (*investapi.SharesResponse, error) {
	return i.tapi.GetAllShares(ctx)
}

func (i *BaseInfoSrv) GetInstrumentInfoByFigi(figi string, ctx context.Context) (*dtotapi.InstrumentResponse, error) {
	req := dtotapi.InstrumentRequest{
		IdType: dtotapi.InstrumentIdTypeFigi,
		Id:     figi,
	}
	return i.tapi.GetInstrumentInfo(&req, ctx)
}

func (i *BaseInfoSrv) GetLastPrices(figis []string, ctx context.Context) (*dtotapi.LastPricesResponse, error) {
	req := dtotapi.LastPricesRequest{Figis: figis}
	return i.tapi.GetLastPrices(&req, ctx)
}
