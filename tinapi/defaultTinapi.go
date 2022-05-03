package tinapi

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"invest-robot/domain"
	"invest-robot/helper"
	investapi "invest-robot/tapigen"
	"time"
)

type DefaultTinApi struct {
	mcl  investapi.MarketDataServiceClient
	mcls investapi.MarketDataStreamServiceClient
}

func NewTinApi() TinApi {
	return DefaultTinApi{
		investapi.NewMarketDataServiceClient(helper.GetClient()),
		investapi.NewMarketDataStreamServiceClient(helper.GetClient()),
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

func (t DefaultTinApi) GetHistory(figis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time) ([]domain.History, error) {
	var resps = make([]domain.History, 0, len(figis))
	ctx := contextWithAuth(context.Background())
	for _, figi := range figis {
		req := investapi.GetCandlesRequest{
			Figi:     figi,
			From:     timestamppb.New(startDate),
			To:       timestamppb.New(endDate),
			Interval: ivl,
		}
		data, err := t.mcl.GetCandles(ctx, &req)
		for _, cndl := range data.GetCandles() {
			histRec := domain.FromHistoricCandle(cndl)
			histRec.Figi = figi
			resps = append(resps, histRec)
		}
		if err != nil {
			return nil, err
		}
	}

	return resps, nil
}

func contextWithAuth(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{"Authorization": "Bearer " + helper.GetTinToken()})
	return metadata.NewOutgoingContext(ctx, md)
}

func (t DefaultTinApi) GetDataStream() (*investapi.MarketDataStreamService_MarketDataStreamClient, error) {
	stream, err := t.mcls.MarketDataStream(context.Background())
	if err != nil {
		return nil, err
	}
	return &stream, nil
}
