package avr

import (
	"github.com/shopspring/decimal"
	"invest-robot/collections"
	"invest-robot/errors"
	"log"
	"time"
)

type DataProc interface {
	GetDataStream() (<-chan procData, error)
	Go()
	Stop() error
}

type procData struct {
	Figi  string
	Time  time.Time
	LAV   decimal.Decimal //average by long window
	SAV   decimal.Decimal //average by short window
	Price decimal.Decimal //current price
}

const (
	ShortDur string = "short_dur"
	LongDur  string = "long_dur"
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
