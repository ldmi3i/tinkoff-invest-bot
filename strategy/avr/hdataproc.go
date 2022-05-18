package avr

import (
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"invest-robot/collections"
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"strconv"
	"time"
)

type DbDataProc struct {
	params map[string]string
	figis  []string
	rep    repository.HistoryRepository
	stopCh chan bool
	hist   []domain.History
	dtCh   chan procData
	logger *zap.SugaredLogger

	sav collections.TList[decimal.Decimal]
	lav collections.TList[decimal.Decimal]
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
	d.sav = collections.NewTList[decimal.Decimal](time.Duration(shortDur) * time.Second)
	d.lav = collections.NewTList[decimal.Decimal](time.Duration(longDur) * time.Second)
	return d.dtCh, nil
}

func (d *DbDataProc) Go() {
	go d.procBg()
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
		d.logger.Debugf("Processing data %+v", hDat)
		sPop := d.sav.Append(hDat.Close, hDat.Time)
		lPop := d.lav.Append(hDat.Close, hDat.Time)
		sOk = sOk || sPop
		lOk = lOk || lPop
		if sOk && lOk {
			sav, err := calcAvg(&d.sav)
			if err != nil {
				d.logger.Errorf("Error while calculating short average:\n%s", err)
				break
			}
			lav, err := calcAvg(&d.lav)
			if err != nil {
				d.logger.Errorf("Error while calculating long average:\n%s", err)
				break
			}
			dat := procData{
				Figi:  hDat.Figi,
				Time:  hDat.Time,
				LAV:   *lav,
				SAV:   *sav,
				Price: hDat.Close,
			}
			d.logger.Debugf("Sending data: %+v", dat)
			d.dtCh <- dat
			time.Sleep(1 * time.Millisecond) //To provide time for mockTrader to finish operation
		}
	}
}

func (d *DbDataProc) Stop() error {
	return errors.NewNotImplemented()
}

func newHistoryDataProc(req *domain.Algorithm, rep repository.HistoryRepository, logger *zap.SugaredLogger) (DataProc, error) {
	return &DbDataProc{
		params: domain.ParamsToMap(req.Params),
		figis:  req.Figis,
		rep:    rep,
		stopCh: make(chan bool, 1),
		dtCh:   make(chan procData),
		logger: logger,
	}, nil
}
