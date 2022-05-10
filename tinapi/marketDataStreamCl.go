package tinapi

import (
	"invest-robot/dto/tapi"
	"invest-robot/errors"
	investapi "invest-robot/tapigen"
	"log"
	"sync/atomic"
)

type MarketDataStreamClient struct {
	client       investapi.MarketDataStreamService_MarketDataStreamClient
	isSubscribed atomic.Value
}

func (c *MarketDataStreamClient) SubscribeCandles(req *tapi.SubscribeCandlesRequest) error {
	instruments := make([]*investapi.CandleInstrument, 0, len(req.Instruments))
	for _, instrument := range req.Instruments {
		instr := investapi.CandleInstrument{
			Figi:     instrument.Figi,
			Interval: investapi.SubscriptionInterval(instrument.Interval),
		}
		instruments = append(instruments, &instr)
	}
	body := investapi.SubscribeCandlesRequest{
		SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
		Instruments:        instruments,
	}
	iReq := investapi.MarketDataRequest{
		Payload: &investapi.MarketDataRequest_SubscribeCandlesRequest{
			SubscribeCandlesRequest: &body,
		},
	}
	err := c.client.Send(&iReq)
	if err != nil {
		log.Printf("Error while subscribing to stream: %s", err)
		return err
	}
	c.isSubscribed.CompareAndSwap(0, 1)
	return nil
}

func (c *MarketDataStreamClient) UnsubscribeAndCloseCandles() error {
	body := investapi.SubscribeCandlesRequest{
		SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_UNSUBSCRIBE,
	}
	req := investapi.MarketDataRequest{
		Payload: &investapi.MarketDataRequest_SubscribeCandlesRequest{
			SubscribeCandlesRequest: &body,
		},
	}
	err := c.client.Send(&req)
	if err != nil {
		log.Printf("Error while sending unsubscribe request to stream:\n%s", err)
		return err
	}
	if err := c.client.CloseSend(); err != nil {
		return err
	}
	return nil
}

func (c *MarketDataStreamClient) RecvCandlesSkipOther() (*tapi.StreamCandleResponse, error) {
	if c.isSubscribed.Load() == 0 {
		return nil, errors.NewUnexpectedError("Receive called before subscribe...")
	}
	recv, err := c.client.Recv()
	if err != nil {
		return nil, err
	}
	candle := recv.GetCandle()
	for candle == nil {
		log.Println("Skipping request...")
		recv, err = c.client.Recv()
		if err != nil {
			return nil, err
		}
		candle = recv.GetCandle()
	}
	return tapi.StreamCandleResponseToDto(candle), nil
}
