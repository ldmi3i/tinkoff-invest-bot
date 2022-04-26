package tinapi

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"invest-robot/helper"
	investapi "invest-robot/tapigen"
	"time"
)

type DefaultTinApi struct {
	mcl investapi.MarketDataServiceClient
}

func NewTinApi() TinApi {
	return DefaultTinApi{
		investapi.NewMarketDataServiceClient(helper.GetClient()),
	}
}

func (t DefaultTinApi) GetOrderBook() (*investapi.GetOrderBookResponse, error) {
	req := investapi.GetOrderBookRequest{
		Figi:  "BBG00T22WKV5",
		Depth: 1,
	}

	ctx := contextWithAuth(context.Background())
	book, err := t.mcl.GetOrderBook(ctx, &req)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (t DefaultTinApi) GetHistory(figis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time) ([]*investapi.GetCandlesResponse, error) {
	var resps = make([]*investapi.GetCandlesResponse, 0, len(figis))
	ctx := contextWithAuth(context.Background())
	for _, figi := range figis {
		req := investapi.GetCandlesRequest{
			Figi:     figi,
			From:     timestamppb.New(startDate),
			To:       timestamppb.New(endDate),
			Interval: ivl,
		}
		data, err := t.mcl.GetCandles(ctx, &req)
		if err != nil {
			return nil, err
		}
		resps = append(resps, data)
	}

	return resps, nil
}

func contextWithAuth(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{"Authorization": "Bearer " + helper.GetTinToken()})
	return metadata.NewOutgoingContext(ctx, md)
}
