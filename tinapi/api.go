package tinapi

import (
	"context"
	"go.uber.org/zap"
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
	GetHistorySorted(figis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time, ctx context.Context) ([]domain.History, error)
	MarketDataStream(ctx context.Context) (investapi.MarketDataStreamService_MarketDataStreamClient, error)
	GetAllShares(ctx context.Context) (*investapi.SharesResponse, error)
	GetInstrumentInfo(req *tapi.InstrumentRequest, ctx context.Context) (*tapi.InstrumentResponse, error)
	GetLastPrices(req *tapi.LastPricesRequest, ctx context.Context) (*tapi.LastPricesResponse, error)
	GetOrderStream(accounts []string, ctx context.Context) (investapi.OrdersStreamService_TradesStreamClient, error)

	PostSandboxOrder(req *tapi.PostOrderRequest, ctx context.Context) (*tapi.PostOrderResponse, error)
	PostProdOrder(req *tapi.PostOrderRequest, ctx context.Context) (*tapi.PostOrderResponse, error)

	CancelSandboxOrder(req *tapi.CancelOrderRequest, ctx context.Context) (*tapi.CancelOrderResponse, error)
	CancelProdOrder(req *tapi.CancelOrderRequest, ctx context.Context) (*tapi.CancelOrderResponse, error)

	GetSandboxOrderState(req *tapi.GetOrderStateRequest, ctx context.Context) (*tapi.GetOrderStateResponse, error)
	GetProdOrderState(req *tapi.GetOrderStateRequest, ctx context.Context) (*tapi.GetOrderStateResponse, error)

	GetSandboxPositions(req *tapi.PositionsRequest, ctx context.Context) (*tapi.PositionsResponse, error)
	GetProdPositions(req *tapi.PositionsRequest, ctx context.Context) (*tapi.PositionsResponse, error)
}

type DefaultTinApi struct {
	marketDatCl   investapi.MarketDataServiceClient
	marketDatStCl investapi.MarketDataStreamServiceClient
	instrCl       investapi.InstrumentsServiceClient
	sandboxCl     investapi.SandboxServiceClient
	orderCl       investapi.OrdersServiceClient
	operationsCl  investapi.OperationsServiceClient
	orderStCl     investapi.OrdersStreamServiceClient
	logger        *zap.SugaredLogger
}

func NewTinApi(logger *zap.SugaredLogger) Api {
	return &DefaultTinApi{
		investapi.NewMarketDataServiceClient(helper.GetClient()),
		investapi.NewMarketDataStreamServiceClient(helper.GetClient()),
		investapi.NewInstrumentsServiceClient(helper.GetClient()),
		investapi.NewSandboxServiceClient(helper.GetClient()),
		investapi.NewOrdersServiceClient(helper.GetClient()),
		investapi.NewOperationsServiceClient(helper.GetClient()),
		investapi.NewOrdersStreamServiceClient(helper.GetClient()),
		logger,
	}
}

func (t *DefaultTinApi) GetHistorySorted(figis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time, ctx context.Context) ([]domain.History, error) {
	var resps = make([]domain.History, 0, len(figis))
	ctxA := contextWithAuth(ctx)
	for _, figi := range figis {
		req := investapi.GetCandlesRequest{
			Figi:     figi,
			From:     timestamppb.New(startDate),
			To:       timestamppb.New(endDate),
			Interval: ivl,
		}
		data, err := t.marketDatCl.GetCandles(ctxA, &req)
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

func (t *DefaultTinApi) MarketDataStream(ctx context.Context) (investapi.MarketDataStreamService_MarketDataStreamClient, error) {
	ctxA := contextWithAuth(ctx)
	stream, err := t.marketDatStCl.MarketDataStream(ctxA)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (t *DefaultTinApi) GetAllShares(ctx context.Context) (*investapi.SharesResponse, error) {
	ctxA := contextWithAuth(ctx)
	req := investapi.InstrumentsRequest{InstrumentStatus: investapi.InstrumentStatus_INSTRUMENT_STATUS_BASE}
	shares, err := t.instrCl.Shares(ctxA, &req)
	if err != nil {
		return nil, err
	}
	return shares, nil
}

func (t *DefaultTinApi) GetInstrumentInfo(req *tapi.InstrumentRequest, ctx context.Context) (*tapi.InstrumentResponse, error) {
	ctxA := contextWithAuth(ctx)
	instrInfo, err := t.instrCl.GetInstrumentBy(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.InstrumentResponseToDto(instrInfo), nil
}

func (t *DefaultTinApi) GetLastPrices(req *tapi.LastPricesRequest, ctx context.Context) (*tapi.LastPricesResponse, error) {
	ctxA := contextWithAuth(ctx)
	prices, err := t.marketDatCl.GetLastPrices(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.LastPricesResponseToDto(prices), nil
}

func (t *DefaultTinApi) PostSandboxOrder(req *tapi.PostOrderRequest, ctx context.Context) (*tapi.PostOrderResponse, error) {
	ctxA := contextWithAuth(ctx)
	log.Println("Post order request:", req.ToTinApi())
	order, err := t.sandboxCl.PostSandboxOrder(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.PostOrderResponseToDto(order), nil
}

func (t *DefaultTinApi) GetSandboxOrderState(req *tapi.GetOrderStateRequest, ctx context.Context) (*tapi.GetOrderStateResponse, error) {
	ctxA := contextWithAuth(ctx)
	resp, err := t.sandboxCl.GetSandboxOrderState(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.OrderStateResponseToDto(resp), nil
}

func (t *DefaultTinApi) GetProdOrderState(req *tapi.GetOrderStateRequest, ctx context.Context) (*tapi.GetOrderStateResponse, error) {
	ctxA := contextWithAuth(ctx)
	resp, err := t.orderCl.GetOrderState(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.OrderStateResponseToDto(resp), nil
}

func (t *DefaultTinApi) PostProdOrder(req *tapi.PostOrderRequest, ctx context.Context) (*tapi.PostOrderResponse, error) {
	ctxA := contextWithAuth(ctx)
	t.logger.Infof("Posting prod order %+v", req.ToTinApi())
	order, err := t.orderCl.PostOrder(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.PostOrderResponseToDto(order), nil
}

func (t *DefaultTinApi) CancelSandboxOrder(req *tapi.CancelOrderRequest, ctx context.Context) (*tapi.CancelOrderResponse, error) {
	ctxA := contextWithAuth(ctx)
	resp, err := t.sandboxCl.CancelSandboxOrder(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.CancelOrderResponseToDto(resp), nil
}

func (t *DefaultTinApi) CancelProdOrder(req *tapi.CancelOrderRequest, ctx context.Context) (*tapi.CancelOrderResponse, error) {
	ctxA := contextWithAuth(ctx)
	resp, err := t.orderCl.CancelOrder(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.CancelOrderResponseToDto(resp), nil
}

func (t *DefaultTinApi) GetSandboxPositions(req *tapi.PositionsRequest, ctx context.Context) (*tapi.PositionsResponse, error) {
	ctxA := contextWithAuth(ctx)
	positions, err := t.sandboxCl.GetSandboxPositions(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.PositionsResponseToDto(positions), nil
}

func (t *DefaultTinApi) GetProdPositions(req *tapi.PositionsRequest, ctx context.Context) (*tapi.PositionsResponse, error) {
	ctxA := contextWithAuth(ctx)
	positions, err := t.operationsCl.GetPositions(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return tapi.PositionsResponseToDto(positions), nil
}

func (t *DefaultTinApi) GetOrderStream(accounts []string, ctx context.Context) (investapi.OrdersStreamService_TradesStreamClient, error) {
	ctxA := contextWithAuth(ctx)
	req := investapi.TradesStreamRequest{Accounts: accounts}
	return t.orderStCl.TradesStream(ctxA, &req)
}
