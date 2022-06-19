package avr

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/collections"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/convert"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/env"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/errors"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/service"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/trade/trmodel"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"strconv"
	"time"
)

type DataProcProd struct {
	stream   investapi.MarketDataStreamService_MarketDataStreamClient
	algo     *entity.Algorithm
	infoSrv  service.InfoSrv
	algoId   uint                               //Algorithm id
	params   map[string]string                  //Algorithm parameters - sizes of AVR windows
	figis    []string                           //List of instrument figis to send to algorithm
	ctx      context.Context                    //Data processor context to save and restore, TODO not implemented
	dtCh     chan procData                      //Channel for algorithm with processed (calculated AVR etc) data from stream
	origDtCh chan *investapi.MarketDataResponse //Channel populated from market stream
	trackId  string                             //Track id in case of incorrect response
	retryMin int                                //Minutes before consecutive retries when data stream broken
	retryNum int                                //Number restore retries when data stream broken

	longDur    int //Extracted to state because of using in extract history method
	savMap     map[string]*collections.TList[decimal.Decimal]
	prevSavMap map[string]trmodel.Timed[decimal.Decimal]
	lavMap     map[string]*collections.TList[decimal.Decimal]
	logger     *zap.SugaredLogger
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
		d.prevSavMap[figi] = trmodel.Timed[decimal.Decimal]{decimal.Zero, time.Now()}
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
			prevSav := d.prevSavMap[candle.Figi]
			price := convert.QuotationToDec(candle.Close)
			dTime := time.Now()
			savL.Append(price, dTime)
			lavL.Append(price, dTime)

			sav, err := calcAvr(savL)
			if err != nil {
				d.logger.Errorf("Error while calculating short average %d: %s", d.algoId, err)
				d.logger.Debugf("Wrong short average: %s", lavL)
				break
			}
			lav, err := calcAvr(lavL)
			if err != nil {
				d.logger.Errorf("Error while calculating long average %d: %s", d.algoId, err)
				d.logger.Debugf("Wrong long average: %s", lavL)
				break
			}
			savDiff := sav.Sub(prevSav.Data)
			timeDiff := dTime.Sub(prevSav.Time).Minutes()
			var derivative decimal.Decimal
			if timeDiff == 0 {
				d.logger.Warnf("No time difference with previous value! Curr: %s, %s; Previous: %s, %s;",
					prevSav.Time, prevSav.Data, dTime, sav)
				derivative = decimal.Zero
			} else {
				derivative = savDiff.Div(decimal.NewFromFloat(timeDiff))
			}

			dat := procData{
				Figi:  candle.Figi,
				Time:  dTime,
				LAV:   lav,
				SAV:   sav,
				DER:   derivative,
				Price: price,
			}
			d.prevSavMap[candle.Figi] = trmodel.Timed[decimal.Decimal]{sav, dTime}
			d.logger.Debugf("Sending data for alg %d: %+v", d.algoId, dat)
			d.dtCh <- dat
		case <-d.ctx.Done():
			d.logger.Info("Algorithm canceling context signal received...")
			break OUT
		}
	}
}

//Background task which receives data from stream and send it to channel (for simpler support of context and processor stopping)
func (d *DataProcProd) processDataInBg() {
	defer func() {
		d.logger.Info("Closing tin API data channel...")
		close(d.origDtCh)
	}()
	for {
		cDat, err := d.stream.Recv()
		if err != nil {
			//Process end of stream and no need to restore
			if err == io.EOF {
				d.logger.Info("Received end of stream...")
				return
			}

			//Process cancel error when context canceled and no need to restore
			if code := status.Code(err); code == codes.Canceled {
				d.logger.Info("Received canceled code: ", code)
				return
			}

			//Trying to restore stream with set number of iterations
			err = d.restoreDataStreamWithRetries()
			if err != nil {
				d.logger.Error(err)
				return
			}
		}
		d.origDtCh <- cDat
	}
}

//Retries to create and subscribe to new channel to restore data receiving
//Uses retryNum and retryMin parameters
func (d *DataProcProd) restoreDataStreamWithRetries() error {
	var err error
	retryNumCp := d.retryNum //Copy to make no change of original setting
	d.logger.Info("Starting stream restoring for alg ", d.algoId, " with ", retryNumCp, " retries and interval (min) ", d.retryMin)
	for ; retryNumCp > 0; retryNumCp-- {
		time.Sleep(time.Duration(d.retryMin) * time.Minute)
		d.logger.Infof("Retry remains %d. Trying to re-create channel after %d min", retryNumCp, d.retryMin)
		d.stream, err = d.infoSrv.GetDataStream(d.ctx)
		if err != nil {
			d.logger.Error("Recreate stream failed with error: ", err)
			continue
		}
		err = d.subscribe()
		if err != nil {
			d.logger.Error("Subscribe new stream failed with error: ", err)
		} else {
			d.logger.Info("Stream successfully restored for algorithm ", d.algoId)
			return nil
		}
	}
	return errors.NewUnexpectedError("Not possible to restore data stream, exiting...")
}

//prefetchHistory populates average windows with data from history to start working effective immediate
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

func newDataProc(req *entity.Algorithm, infoSrv service.InfoSrv, logger *zap.SugaredLogger) (DataProc, error) {
	env.GetDbUser()
	return &DataProcProd{
		algo:       req,
		infoSrv:    infoSrv,
		algoId:     req.ID,
		params:     entity.ParamsToMap(req.Params),
		figis:      req.Figis,
		dtCh:       make(chan procData),
		origDtCh:   make(chan *investapi.MarketDataResponse),
		savMap:     make(map[string]*collections.TList[decimal.Decimal]),
		prevSavMap: make(map[string]trmodel.Timed[decimal.Decimal]),
		lavMap:     make(map[string]*collections.TList[decimal.Decimal]),
		logger:     logger,
		retryMin:   env.GetRetryMin(),
		retryNum:   env.GetRetryNum(),
	}, nil
}
