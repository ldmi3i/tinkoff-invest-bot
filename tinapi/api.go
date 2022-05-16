package tinapi

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"invest-robot/domain"
	"invest-robot/dto/tapi"
	"invest-robot/helper"
	investapi "invest-robot/tapigen"
	"log"
	"time"
)

//Api is a wrapper under generated GRPC to provide only required methods
type Api interface {
	GetOrderBook() (*investapi.GetOrderBookResponse, error)
	GetHistorySorted(figis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time) ([]domain.History, error)
	MarketDataStream() (investapi.MarketDataStreamService_MarketDataStreamClient, error)
	GetAllShares() (*investapi.SharesResponse, error)
	GetInstrumentInfo(req *tapi.InstrumentRequest) (*tapi.InstrumentResponse, error)
	GetLastPrices(req *tapi.LastPricesRequest) (*tapi.LastPricesResponse, error)
	GetOrderStream(accounts []string) (investapi.OrdersStreamService_TradesStreamClient, error)

	PostSandboxOrder(req *tapi.PostOrderRequest) (*tapi.PostOrderResponse, error)
	PostProdOrder(req *tapi.PostOrderRequest) (*tapi.PostOrderResponse, error)

	CancelSandboxOrder(req *tapi.CancelOrderRequest) (*tapi.CancelOrderResponse, error)
	CancelProdOrder(req *tapi.CancelOrderRequest) (*tapi.CancelOrderResponse, error)

	GetSandboxOrderState(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error)
	GetProdOrderState(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error)

	GetSandboxPositions(req *tapi.PositionsRequest) (*tapi.PositionsResponse, error)
	GetProdPositions(req *tapi.PositionsRequest) (*tapi.PositionsResponse, error)
}

type DefaultTinApi struct {
	marketDatCl   investapi.MarketDataServiceClient
	marketDatStCl investapi.MarketDataStreamServiceClient
	instrCl       investapi.InstrumentsServiceClient
	sandboxCl     investapi.SandboxServiceClient
	orderCl       investapi.OrdersServiceClient
	operationsCl  investapi.OperationsServiceClient
	orderStCl     investapi.OrdersStreamServiceClient
}

func NewTinApi() Api {
	return &DefaultTinApi{
		investapi.NewMarketDataServiceClient(helper.GetClient()),
		investapi.NewMarketDataStreamServiceClient(helper.GetClient()),
		investapi.NewInstrumentsServiceClient(helper.GetClient()),
		investapi.NewSandboxServiceClient(helper.GetClient()),
		investapi.NewOrdersServiceClient(helper.GetClient()),
		investapi.NewOperationsServiceClient(helper.GetClient()),
		investapi.NewOrdersStreamServiceClient(helper.GetClient()),
	}
}

func (t *DefaultTinApi) GetOrderBook() (*investapi.GetOrderBookResponse, error) {
	req := investapi.GetOrderBookRequest{
		Figi:  "BBG00T22WKV5",
		Depth: 1,
	}

	ctx := contextWithAuth(context.Background())
	book, err := t.marketDatCl.GetOrderBook(ctx, &req)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (t *DefaultTinApi) GetHistorySorted(figis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time) ([]domain.History, error) {
	var resps = make([]domain.History, 0, len(figis))
	ctx := contextWithAuth(context.Background())
	for _, figi := range figis {
		req := investapi.GetCandlesRequest{
			Figi:     figi,
			From:     timestamppb.New(startDate),
			To:       timestamppb.New(endDate),
			Interval: ivl,
		}
		data, err := t.marketDatCl.GetCandles(ctx, &req)
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
	md := metadata.New(map[string]string{
		"Authorization": "Bearer " + helper.GetTinToken(),
		"x-app-name":    "ldmi3i",
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func (t *DefaultTinApi) MarketDataStream() (investapi.MarketDataStreamService_MarketDataStreamClient, error) {
	ctx := contextWithAuth(context.Background())
	stream, err := t.marketDatStCl.MarketDataStream(ctx)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (t *DefaultTinApi) GetAllShares() (*investapi.SharesResponse, error) {
	ctx := contextWithAuth(context.Background())
	req := investapi.InstrumentsRequest{InstrumentStatus: investapi.InstrumentStatus_INSTRUMENT_STATUS_BASE}
	shares, err := t.instrCl.Shares(ctx, &req)
	if err != nil {
		return nil, err
	}
	return shares, nil
}

func (t *DefaultTinApi) GetInstrumentInfo(req *tapi.InstrumentRequest) (*tapi.InstrumentResponse, error) {
	ctx := contextWithAuth(context.Background())
	instrInfo, err := t.instrCl.GetInstrumentBy(ctx, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.InstrumentResponseToDto(instrInfo), nil
}

func (t *DefaultTinApi) GetLastPrices(req *tapi.LastPricesRequest) (*tapi.LastPricesResponse, error) {
	ctx := contextWithAuth(context.Background())
	prices, err := t.marketDatCl.GetLastPrices(ctx, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.LastPricesResponseToDto(prices), nil
}

func (t *DefaultTinApi) PostSandboxOrder(req *tapi.PostOrderRequest) (*tapi.PostOrderResponse, error) {
	ctx := contextWithAuth(context.Background())
	log.Println("Post order request:", req.ToTinApi())
	order, err := t.sandboxCl.PostSandboxOrder(ctx, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.PostOrderResponseToDto(order), nil
}

func (t *DefaultTinApi) GetSandboxOrderState(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error) {
	ctx := contextWithAuth(context.Background())
	resp, err := t.sandboxCl.GetSandboxOrderState(ctx, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.OrderStateResponseToDto(resp), nil
}

func (t *DefaultTinApi) GetProdOrderState(req *tapi.GetOrderStateRequest) (*tapi.GetOrderStateResponse, error) {
	ctx := contextWithAuth(context.Background())
	resp, err := t.orderCl.GetOrderState(ctx, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.OrderStateResponseToDto(resp), nil
}

func (t *DefaultTinApi) PostProdOrder(req *tapi.PostOrderRequest) (*tapi.PostOrderResponse, error) {
	ctx := contextWithAuth(context.Background())
	order, err := t.orderCl.PostOrder(ctx, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.PostOrderResponseToDto(order), nil
}

func (t *DefaultTinApi) CancelSandboxOrder(req *tapi.CancelOrderRequest) (*tapi.CancelOrderResponse, error) {
	ctx := contextWithAuth(context.Background())
	resp, err := t.sandboxCl.CancelSandboxOrder(ctx, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.CancelOrderResponseToDto(resp), nil
}

func (t *DefaultTinApi) CancelProdOrder(req *tapi.CancelOrderRequest) (*tapi.CancelOrderResponse, error) {
	ctx := contextWithAuth(context.Background())
	resp, err := t.orderCl.CancelOrder(ctx, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.CancelOrderResponseToDto(resp), nil
}

func (t *DefaultTinApi) GetSandboxPositions(req *tapi.PositionsRequest) (*tapi.PositionsResponse, error) {
	ctx := contextWithAuth(context.Background())
	positions, err := t.sandboxCl.GetSandboxPositions(ctx, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.PositionsResponseToDto(positions), nil
}

func (t *DefaultTinApi) GetProdPositions(req *tapi.PositionsRequest) (*tapi.PositionsResponse, error) {
	ctx := contextWithAuth(context.Background())
	positions, err := t.operationsCl.GetPositions(ctx, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.PositionsResponseToDto(positions), nil
}

func (t *DefaultTinApi) GetOrderStream(accounts []string) (investapi.OrdersStreamService_TradesStreamClient, error) {
	ctx := contextWithAuth(context.Background())
	req := investapi.TradesStreamRequest{Accounts: accounts}
	return t.orderStCl.TradesStream(ctx, &req)
}
