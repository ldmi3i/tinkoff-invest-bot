package tinapi

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/connections/grpc"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto/dtotapi"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/env"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"
)

//Api is a wrapper under generated GRPC to provide only required methods
type Api interface {
	GetHistory(figis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time, ctx context.Context) ([]entity.History, error)
	MarketDataStream(ctx context.Context) (investapi.MarketDataStreamService_MarketDataStreamClient, error)
	GetAllShares(ctx context.Context) (*dtotapi.SharesResponse, error)
	GetInstrumentInfo(req *dtotapi.InstrumentRequest, ctx context.Context) (*dtotapi.InstrumentResponse, error)
	GetLastPrices(req *dtotapi.LastPricesRequest, ctx context.Context) (*dtotapi.LastPricesResponse, error)
	GetOrderStream(accounts []string, ctx context.Context) (investapi.OrdersStreamService_TradesStreamClient, error)

	PostSandboxOrder(req *dtotapi.PostOrderRequest, ctx context.Context) (*dtotapi.PostOrderResponse, error)
	PostProdOrder(req *dtotapi.PostOrderRequest, ctx context.Context) (*dtotapi.PostOrderResponse, error)

	CancelSandboxOrder(req *dtotapi.CancelOrderRequest, ctx context.Context) (*dtotapi.CancelOrderResponse, error)
	CancelProdOrder(req *dtotapi.CancelOrderRequest, ctx context.Context) (*dtotapi.CancelOrderResponse, error)

	GetSandboxOrderState(req *dtotapi.OrderStateRequest, ctx context.Context) (*dtotapi.OrderStateResponse, error)
	GetProdOrderState(req *dtotapi.OrderStateRequest, ctx context.Context) (*dtotapi.OrderStateResponse, error)

	GetSandboxPositions(req *dtotapi.PositionsRequest, ctx context.Context) (*dtotapi.PositionsResponse, error)
	GetProdPositions(req *dtotapi.PositionsRequest, ctx context.Context) (*dtotapi.PositionsResponse, error)

	GetSandboxAccounts(ctx context.Context) (*dtotapi.AccountsResponse, error)
	GetProdAccounts(ctx context.Context) (*dtotapi.AccountsResponse, error)
}

type DefaultTinApi struct {
	marketDatCl   investapi.MarketDataServiceClient
	marketDatStCl investapi.MarketDataStreamServiceClient
	instrCl       investapi.InstrumentsServiceClient
	sandboxCl     investapi.SandboxServiceClient
	orderCl       investapi.OrdersServiceClient
	operationsCl  investapi.OperationsServiceClient
	orderStCl     investapi.OrdersStreamServiceClient
	usersCl       investapi.UsersServiceClient
	logger        *zap.SugaredLogger
}

func NewTinApi(logger *zap.SugaredLogger) Api {
	return &DefaultTinApi{
		investapi.NewMarketDataServiceClient(grpc.GetClient()),
		investapi.NewMarketDataStreamServiceClient(grpc.GetClient()),
		investapi.NewInstrumentsServiceClient(grpc.GetClient()),
		investapi.NewSandboxServiceClient(grpc.GetClient()),
		investapi.NewOrdersServiceClient(grpc.GetClient()),
		investapi.NewOperationsServiceClient(grpc.GetClient()),
		investapi.NewOrdersStreamServiceClient(grpc.GetClient()),
		investapi.NewUsersServiceClient(grpc.GetClient()),
		logger,
	}
}

func (t *DefaultTinApi) GetHistory(figis []string, ivl investapi.CandleInterval, startDate time.Time, endDate time.Time, ctx context.Context) ([]entity.History, error) {
	var resps = make([]entity.History, 0, len(figis))
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
			histRec := entity.FromHistoricCandle(cndl)
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
		"Authorization": "Bearer " + env.GetTinToken(),
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

func (t *DefaultTinApi) GetAllShares(ctx context.Context) (*dtotapi.SharesResponse, error) {
	ctxA := contextWithAuth(ctx)
	req := investapi.InstrumentsRequest{InstrumentStatus: investapi.InstrumentStatus_INSTRUMENT_STATUS_BASE}
	shares, err := t.instrCl.Shares(ctxA, &req)
	if err != nil {
		return nil, err
	}
	return dtotapi.SharesResponseToDto(shares), nil
}

func (t *DefaultTinApi) GetInstrumentInfo(req *dtotapi.InstrumentRequest, ctx context.Context) (*dtotapi.InstrumentResponse, error) {
	ctxA := contextWithAuth(ctx)
	instrInfo, err := t.instrCl.GetInstrumentBy(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return dtotapi.InstrumentResponseToDto(instrInfo), nil
}

func (t *DefaultTinApi) GetLastPrices(req *dtotapi.LastPricesRequest, ctx context.Context) (*dtotapi.LastPricesResponse, error) {
	ctxA := contextWithAuth(ctx)
	prices, err := t.marketDatCl.GetLastPrices(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return dtotapi.LastPricesResponseToDto(prices), nil
}

func (t *DefaultTinApi) PostSandboxOrder(req *dtotapi.PostOrderRequest, ctx context.Context) (*dtotapi.PostOrderResponse, error) {
	ctxA := contextWithAuth(ctx)
	log.Println("Post order request:", req.ToTinApi())
	order, err := t.sandboxCl.PostSandboxOrder(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return dtotapi.PostOrderResponseToDto(order), nil
}

func (t *DefaultTinApi) GetSandboxOrderState(req *dtotapi.OrderStateRequest, ctx context.Context) (*dtotapi.OrderStateResponse, error) {
	ctxA := contextWithAuth(ctx)
	resp, err := t.sandboxCl.GetSandboxOrderState(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return dtotapi.OrderStateResponseToDto(resp), nil
}

func (t *DefaultTinApi) GetProdOrderState(req *dtotapi.OrderStateRequest, ctx context.Context) (*dtotapi.OrderStateResponse, error) {
	ctxA := contextWithAuth(ctx)
	resp, err := t.orderCl.GetOrderState(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return dtotapi.OrderStateResponseToDto(resp), nil
}

func (t *DefaultTinApi) PostProdOrder(req *dtotapi.PostOrderRequest, ctx context.Context) (*dtotapi.PostOrderResponse, error) {
	ctxA := contextWithAuth(ctx)
	t.logger.Infof("Posting prod order %+v", req.ToTinApi())
	order, err := t.orderCl.PostOrder(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return dtotapi.PostOrderResponseToDto(order), nil
}

func (t *DefaultTinApi) CancelSandboxOrder(req *dtotapi.CancelOrderRequest, ctx context.Context) (*dtotapi.CancelOrderResponse, error) {
	ctxA := contextWithAuth(ctx)
	resp, err := t.sandboxCl.CancelSandboxOrder(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return dtotapi.CancelOrderResponseToDto(resp), nil
}

func (t *DefaultTinApi) CancelProdOrder(req *dtotapi.CancelOrderRequest, ctx context.Context) (*dtotapi.CancelOrderResponse, error) {
	ctxA := contextWithAuth(ctx)
	resp, err := t.orderCl.CancelOrder(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return dtotapi.CancelOrderResponseToDto(resp), nil
}

func (t *DefaultTinApi) GetSandboxPositions(req *dtotapi.PositionsRequest, ctx context.Context) (*dtotapi.PositionsResponse, error) {
	ctxA := contextWithAuth(ctx)
	positions, err := t.sandboxCl.GetSandboxPositions(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return dtotapi.PositionsResponseToDto(positions), nil
}

func (t *DefaultTinApi) GetProdPositions(req *dtotapi.PositionsRequest, ctx context.Context) (*dtotapi.PositionsResponse, error) {
	ctxA := contextWithAuth(ctx)
	positions, err := t.operationsCl.GetPositions(ctxA, req.ToTinApi())
	if err != nil {
		return nil, err
	}
	return dtotapi.PositionsResponseToDto(positions), nil
}

func (t *DefaultTinApi) GetOrderStream(accounts []string, ctx context.Context) (investapi.OrdersStreamService_TradesStreamClient, error) {
	ctxA := contextWithAuth(ctx)
	req := investapi.TradesStreamRequest{Accounts: accounts}
	return t.orderStCl.TradesStream(ctxA, &req)
}

func (t *DefaultTinApi) GetSandboxAccounts(ctx context.Context) (*dtotapi.AccountsResponse, error) {
	ctxA := contextWithAuth(ctx)
	accounts, err := t.sandboxCl.GetSandboxAccounts(ctxA, &investapi.GetAccountsRequest{})
	if err != nil {
		return nil, err
	}
	return dtotapi.AccountsResponseToDto(accounts), nil
}

func (t *DefaultTinApi) GetProdAccounts(ctx context.Context) (*dtotapi.AccountsResponse, error) {
	ctxA := contextWithAuth(ctx)
	accounts, err := t.usersCl.GetAccounts(ctxA, &investapi.GetAccountsRequest{})
	if err != nil {
		return nil, err
	}
	return dtotapi.AccountsResponseToDto(accounts), nil
}
