package avr

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/collections"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/errors"
	"github.com/shopspring/decimal"
	"log"
	"time"
)

//DataProc provides data with short and long AVR windows for the algorithm
type DataProc interface {
	//GetDataStream provide channel with data stream from processor
	GetDataStream() (<-chan procData, error)
	//Go commands DataProc to start processing data in background
	Go(ctx context.Context) error
}

type procData struct {
	Figi  string
	Time  time.Time
	LAV   decimal.Decimal //average by long window
	SAV   decimal.Decimal //average by short window
	DER   decimal.Decimal //Short window current derivative
	Price decimal.Decimal //current price
}

//Average window parameters
const (
	ShortDur string = "short_dur" //Short window length in sec
	LongDur  string = "long_dur"  //Long window length in sec
)

func calcAvr(lst *collections.TList[decimal.Decimal]) (avr decimal.Decimal, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("calcAvr method failed and recovered, info: %s", r)
			err = errors.ConvertToError(r)
		}
	}()
	if lst.IsEmpty() {
		log.Println("Requested average of empty list...")
		return decimal.Zero, errors.NewUnexpectedError("requested average calc on empty list")
	}
	var weightSum int64
	var prevWeight int64
	var currWeight int64
	var sum decimal.Decimal
	prevVal := lst.First()
	next := prevVal.Next()
	if next == nil { //One element case
		return prevVal.GetData(), nil //Average equals to value
	}
	prevTime := prevVal.GetTime()
	for ; next != nil; next = next.Next() {
		currTime := next.GetTime()
		currWeight = currTime.Sub(prevTime).Milliseconds() / 2 //Half of interval to previous value
		fullWeight := prevWeight + currWeight                  //Full weight is half of intervals around measured data from both sizes
		weightSum += fullWeight
		//Add previous data weight with left + right half intervals (related interval)
		sum = sum.Add(decimal.NewFromInt(fullWeight).Mul(prevVal.GetData())) //Sum of all weights
		prevVal = next
		prevWeight = currWeight
		prevTime = currTime
	}
	//Take into account half of interval and the last value
	sum = sum.Add(decimal.NewFromInt(prevWeight).Mul(prevVal.GetData()))
	weightSum += prevWeight

	res := sum.Div(decimal.NewFromInt(weightSum))
	return res, nil
}
