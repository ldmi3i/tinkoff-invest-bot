package avr

import (
	"context"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"invest-robot/collections"
	"invest-robot/convert"
	"invest-robot/domain"
	"invest-robot/service"
	investapi "invest-robot/tapigen"
	"io"
	"strconv"
	"time"
)

type DataProcProd struct {
	stream   investapi.MarketDataStreamService_MarketDataStreamClient
	algo     *domain.Algorithm
	infoSrv  service.InfoSrv
	algoId   uint
	params   map[string]string
	figis    []string
	ctx      context.Context
	dtCh     chan procData
	origDtCh chan *investapi.MarketDataResponse
	trackId  string

	longDur int
	savMap  map[string]*collections.TList[decimal.Decimal]
	lavMap  map[string]*collections.TList[decimal.Decimal]
	logger  *zap.SugaredLogger
}

func (d *DataProcProd) GetDataStream() (<-chan procData, error) {
	shortDur, err := strconv.Atoi(d.params[ShortDur])
	if err != nil {
		return nil, err
	}
	d.longDur, err = strconv.Atoi(d.params[LongDur])
	if err != nil {
		return nil, err
	}

	for _, figi := range d.figis {
		sav := collections.NewTList[decimal.Decimal](time.Duration(shortDur) * time.Second)
		lav := collections.NewTList[decimal.Decimal](time.Duration(d.longDur) * time.Second)
		d.savMap[figi] = &sav
		d.lavMap[figi] = &lav
	}
	return d.dtCh, nil
}

func (d *DataProcProd) Go(ctx context.Context) error {
	d.ctx = ctx
	stream, err := d.infoSrv.GetDataStream(d.ctx)
	if err != nil {
		return err
	}
	d.stream = stream
	go d.procBg()
	return nil
}

func (d *DataProcProd) procBg() {
	defer func() {
		close(d.dtCh)
		d.logger.Infof("Data processor stopped, id %d...", d.algoId)
	}()
	err := d.prefetchHistory()
	if err != nil {
		d.logger.Errorf("Error while prefetching history, id %d: %s", d.algoId, err)
		return
	}
	if err = d.subscribe(); err != nil {
		d.logger.Errorf("Error while subsribing to candles, id %d: %s", d.algoId, err)
		return
	}
	go d.processDataInBg()
OUT:
	for {
		select {
		case cDat, ok := <-d.origDtCh:
			if !ok {
				d.logger.Info("Data channel closed, breaking data processor cycle...")
				break OUT
			}
			candle := cDat.GetCandle()
			if candle == nil {
				if cDat.GetPing() == nil {
					d.logger.Infof("Received nil candle, id: %d, tracking id: %s, full response: %+v", d.algoId, d.trackId, cDat)
				}
				continue
			}
			savL, ok := d.savMap[candle.Figi]
			if !ok {
				d.logger.Infof("WARN received figi that not presented in listening list, id: %d", d.algoId)
			}
			lavL := d.lavMap[candle.Figi]
			price := convert.QuotationToDec(candle.Close)
			dTime := candle.Time.AsTime()
			savL.Append(price, dTime)
			lavL.Append(price, dTime)

			sav, err := calcAvg(savL)
			if err != nil {
				d.logger.Errorf("Error while calculating short average %d: %s", d.algoId, err)
				break
			}
			lav, err := calcAvg(lavL)
			if err != nil {
				d.logger.Errorf("Error while calculating long average %d: %s", d.algoId, err)
				break
			}
			dat := procData{
				Figi:  candle.Figi,
				Time:  dTime,
				LAV:   *lav,
				SAV:   *sav,
				Price: price,
			}
			d.logger.Debugf("Sending data for alg %d: %+v", d.algoId, dat)
			d.dtCh <- dat
		case <-d.ctx.Done():
			d.logger.Info("Algorithm canceling context signal received...")
			break OUT
		}
	}
}

func (d *DataProcProd) processDataInBg() {
	defer func() {
		d.logger.Info("Closing tin API data channel...")
		close(d.origDtCh)
	}()
	for {
		cDat, err := d.stream.Recv()
		if err != nil {
			if err == io.EOF {
				d.logger.Info("Received end of stream...")
				return
			}

			//To process errors from grpc stream (Context cancel etc)
			if code := status.Code(err); code == codes.Canceled {
				d.logger.Info("Received canceled code: ", code)
				return
			}

			d.logger.Info("Trying to resubscribe to channel")
			err = d.subscribe()
			if err != nil {
				d.logger.Error("Subscription failed with error: ", err, " Stopping processor...")
				return
			}
		}
		d.origDtCh <- cDat
	}
}

func (d *DataProcProd) prefetchHistory() error {
	endTime := time.Now()
	dur := time.Duration(-d.longDur) * time.Second
	startTime := endTime.Add(dur)
	history, err := d.infoSrv.GetHistorySorted(d.figis, investapi.CandleInterval_CANDLE_INTERVAL_1_MIN, startTime, endTime, d.ctx)
	if err != nil {
		return err
	}
	for _, hRec := range history {
		sav := d.savMap[hRec.Figi]
		lav := d.lavMap[hRec.Figi]
		sav.Append(hRec.Close, hRec.Time)
		lav.Append(hRec.Close, hRec.Time)
	}
	return nil
}

func (d *DataProcProd) subscribe() error {
	d.logger.Info("Subscribing to figis: ", d.figis)
	instruments := make([]*investapi.CandleInstrument, 0, len(d.figis))
	for _, figi := range d.figis {
		instr := investapi.CandleInstrument{
			Figi:     figi,
			Interval: 1,
		}
		instruments = append(instruments, &instr)
	}
	body := investapi.SubscribeCandlesRequest{
		SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
		Instruments:        instruments,
	}
	req := investapi.MarketDataRequest{
		Payload: &investapi.MarketDataRequest_SubscribeCandlesRequest{
			SubscribeCandlesRequest: &body,
		},
	}
	err := d.stream.Send(&req)
	if err != nil {
		d.logger.Errorf("Error while subscribing to stream: %s", err)
		return err
	}
	resp, err := d.stream.Recv()
	if err != nil {
		d.logger.Errorf("Error while awaiting subscription response: %s", err)
		return err
	}
	d.logger.Info("Subscription response received: ", resp)
	d.trackId = resp.GetSubscribeCandlesResponse().GetTrackingId()
	return nil
}

func (d *DataProcProd) unsubscribe() error {
	d.logger.Info("Unsubscribe data processor...")
	body := investapi.SubscribeCandlesRequest{
		SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_UNSUBSCRIBE,
	}
	req := investapi.MarketDataRequest{
		Payload: &investapi.MarketDataRequest_SubscribeCandlesRequest{
			SubscribeCandlesRequest: &body,
		},
	}
	err := d.stream.Send(&req)
	if err != nil {
		d.logger.Errorf("Error while sending unsubscribe request to stream:\n%s", err)
		return err
	}
	if err := d.stream.CloseSend(); err != nil {
		d.logger.Error("Error while sending close: ", err)
		return err
	}
	d.logger.Info("Unsubscribed successfully!")
	return nil
}

func (d *DataProcProd) Stop() error {
	d.logger.Info("Stopping data processor...")
	err := d.unsubscribe()
	if err != nil {
		d.logger.Errorf("Error while unsubscribing to candles, id %d: %s", d.algoId, err)
	}
	d.logger.Info("Stop signal send")
	return nil
}

func newDataProc(req *domain.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (DataProc, error) {
	return &DataProcProd{
		algo:     req,
		infoSrv:  infoSrv,
		algoId:   req.ID,
		params:   domain.ParamsToMap(req.Params),
		figis:    req.Figis,
		dtCh:     make(chan procData),
		origDtCh: make(chan *investapi.MarketDataResponse),
		savMap:   make(map[string]*collections.TList[decimal.Decimal]),
		lavMap:   make(map[string]*collections.TList[decimal.Decimal]),
		logger:   logger,
	}, nil
}
