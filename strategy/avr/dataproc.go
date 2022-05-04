package avr

import (
	"github.com/shopspring/decimal"
	"invest-robot/collections"
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
	"log"
	"strconv"
	"time"
)

type procData struct {
	Figi string
	Time time.Time
	LAV  decimal.Decimal //average by long window
	SAV  decimal.Decimal //average by short window
}

const (
	ShortDur string = "short_dur"
	LongDur  string = "long_dur"
)

type DataProc interface {
	GetDataStream() (<-chan procData, error)
	Go()
	Stop() error
}

type ApiDataProc struct {
	info *service.InfoSrv
}

func (a *ApiDataProc) GetDataStream() (<-chan procData, error) {
	return nil, errors.NewNotImplemented()
}

func (a *ApiDataProc) Go() {
	//todo implement me
}

func (a *ApiDataProc) Stop() error {
	return errors.NewNotImplemented()
}

func newApiDataProc(req domain.Algorithm, infoSrv *service.InfoSrv) (DataProc, error) {
	return &ApiDataProc{infoSrv}, nil
}

type DbDataProc struct {
	params map[string]string
	figis  []string
	rep    *repository.HistoryRepository
	stopCh chan bool
	hist   []domain.History
	dtCh   chan procData

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
	d.hist, err = (*d.rep).FindAllByFigis(d.figis)
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
	log.Printf("Start processing history data, full size: %d", len(d.hist))
	defer func() {
		close(d.dtCh)
		log.Printf("Data processor stopped...")
	}()
	sOk := false
	lOk := false
	for _, hDat := range d.hist {
		log.Printf("Processing data %+v", hDat)
		sPop := d.sav.Append(hDat.Close, hDat.Time)
		lPop := d.lav.Append(hDat.Close, hDat.Time)
		sOk = sOk || sPop
		lOk = lOk || lPop
		if sOk && lOk {
			sav, err := d.calcAvg(d.sav)
			if err != nil {
				log.Printf("Error while calculating short average:\n%s", err)
				break
			}
			lav, err := d.calcAvg(d.lav)
			if err != nil {
				log.Printf("Error while calculating long average:\n%s", err)
				break
			}
			dat := procData{
				Figi: hDat.Figi,
				Time: hDat.Time,
				LAV:  *lav,
				SAV:  *sav,
			}
			log.Printf("Sending data: %+v", dat)
			d.dtCh <- dat
			time.Sleep(time.Millisecond)
		}
	}
}

func (d *DbDataProc) calcAvg(lst collections.TList[decimal.Decimal]) (*decimal.Decimal, error) {
	if lst.IsEmpty() {
		log.Println("Requested average of empty list...")
		return nil, errors.NewUnexpectedError("requested average calc on empty list")
	}
	cnt := 0
	sum := decimal.Zero
	for next := lst.First(); next != nil; next = next.Next() {
		cnt += 1
		sum = sum.Add(next.GetData())
		//log.Printf("Calc data: %s , count: %d", next.GetData(), cnt)
	}
	res := sum.Div(decimal.NewFromInt(int64(cnt)))
	return &res, nil
}

func (d *DbDataProc) Stop() error {
	return errors.NewNotImplemented()
}

func newHistoryDataProc(req domain.Algorithm, rep *repository.HistoryRepository) (DataProc, error) {
	return &DbDataProc{
		params: domain.ParamsToMap(req.Params),
		figis:  req.Figis,
		rep:    rep,
		stopCh: make(chan bool, 1),
		dtCh:   make(chan procData),
	}, nil
}
