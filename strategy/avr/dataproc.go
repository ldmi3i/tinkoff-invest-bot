package avr

import (
	"context"
	"github.com/shopspring/decimal"
	"invest-robot/collections"
	"invest-robot/errors"
	"log"
	"time"
)

//DataProc provides data with short and long AVR windows for the algorithm
type DataProc interface {
	GetDataStream() (<-chan procData, error)
	Go(ctx context.Context) error
}

type procData struct {
	Figi  string
	Time  time.Time
	LAV   decimal.Decimal //average by long window
	SAV   decimal.Decimal //average by short window
	Price decimal.Decimal //current price
}

//Average window parameters
const (
	ShortDur string = "short_dur" //Short window length in sec
	LongDur  string = "long_dur"  //Long window length in sec
)

func calcAvg(lst *collections.TList[decimal.Decimal]) (*decimal.Decimal, error) {
	if lst.IsEmpty() {
		log.Println("Requested average of empty list...")
		return nil, errors.NewUnexpectedError("requested average calc on empty list")
	}
	cnt := 0
	sum := decimal.Zero
	for next := lst.First(); next != nil; next = next.Next() {
		cnt += 1
		sum = sum.Add(next.GetData())
	}
	res := sum.Div(decimal.NewFromInt(int64(cnt)))
	return &res, nil
}
