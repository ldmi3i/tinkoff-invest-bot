package avr

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/collections"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type DbDataProc struct {
	params map[string]string
	figis  []string
	rep    repository.HistoryRepository
	hist   []entity.History
	dtCh   chan procData
	logger *zap.SugaredLogger
	ctx    context.Context

	savMap     map[string]*collections.TList[decimal.Decimal]
	prevSavMap map[string]decimal.Decimal
	lavMap     map[string]*collections.TList[decimal.Decimal]
}

func (d *DbDataProc) GetDataStream() (<-chan procData, error) {
	shortDur, err := strconv.Atoi(d.params[ShortDur])
	if err != nil {
		return nil, err
	}
	longDur, err := strconv.Atoi(d.params[LongDur])
	if err != nil {
		return nil, err
	}
	d.hist, err = d.rep.FindAllByFigis(d.figis)
	if err != nil {
		return nil, err
	}
	for _, figi := range d.figis {
		sav := collections.NewTList[decimal.Decimal](time.Duration(shortDur) * time.Second)
		lav := collections.NewTList[decimal.Decimal](time.Duration(longDur) * time.Second)
		d.prevSavMap[figi] = decimal.Zero
		d.savMap[figi] = &sav
		d.lavMap[figi] = &lav
	}
	return d.dtCh, nil
}

func (d *DbDataProc) Go(ctx context.Context) error {
	d.ctx = ctx
	go d.procBg()
	return nil
}

func (d *DbDataProc) procBg() {
	d.logger.Infof("Start processing history data, full size: %d", len(d.hist))
	defer func() {
		close(d.dtCh)
		d.logger.Infof("Data processor stopped...")
	}()
	sOk := false
	lOk := false
	for _, hDat := range d.hist {
		select {
		case <-d.ctx.Done():
			d.logger.Info("Canceled context, stopping processor...")
			return
		default:
			savL, ok := d.savMap[hDat.Figi]
			if !ok {
				d.logger.Infof("WARN received figi that not presented in listening list, id")
				continue
			}
			lavL := d.lavMap[hDat.Figi]
			prevSav := d.prevSavMap[hDat.Figi]
			//d.logger.Debugf("Processing data %+v", hDat)
			sPop := savL.Append(hDat.Close, hDat.Time)
			lPop := lavL.Append(hDat.Close, hDat.Time)
			sOk = sOk || sPop
			lOk = lOk || lPop
			if sOk && lOk {
				sav, err := calcAvr(savL)
				if err != nil {
					d.logger.Errorf("Error while calculating short average:\n%s", err)
					break
				}
				lav, err := calcAvr(lavL)
				if err != nil {
					d.logger.Errorf("Error while calculating long average:\n%s", err)
					break
				}
				dat := procData{
					Figi:  hDat.Figi,
					Time:  hDat.Time,
					LAV:   lav,
					SAV:   sav,
					DER:   sav.Sub(prevSav).Mul(decimal.NewFromInt(int64(savL.GetSize()))),
					Price: hDat.Close,
				}
				d.prevSavMap[hDat.Figi] = sav
				d.logger.Debugf("Sending data: %+v", dat)
				d.dtCh <- dat
				time.Sleep(1 * time.Millisecond) //To provide time for mockTrader to finish operation
			}
		}
	}
}

func newHistoryDataProc(req *entity.Algorithm, rep repository.HistoryRepository, logger *zap.SugaredLogger) (DataProc, error) {
	return &DbDataProc{
		params:     entity.ParamsToMap(req.Params),
		figis:      req.Figis,
		rep:        rep,
		dtCh:       make(chan procData),
		savMap:     make(map[string]*collections.TList[decimal.Decimal]),
		prevSavMap: make(map[string]decimal.Decimal),
		lavMap:     make(map[string]*collections.TList[decimal.Decimal]),
		logger:     logger,
	}, nil
}
