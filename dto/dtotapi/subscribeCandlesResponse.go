package dtotapi

import (
	"github.com/shopspring/decimal"
	"invest-robot/convert"
	investapi "invest-robot/tapigen"
	"time"
)

type StreamCandleResponse struct {
	Figi      string
	Interval  int
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    int64
	TimeStart time.Time
	TimeEnd   time.Time
}

func StreamCandleResponseToDto(resp *investapi.Candle) *StreamCandleResponse {
	if resp == nil {
		return nil
	}
	return &StreamCandleResponse{
		Figi:      resp.Figi,
		Interval:  int(resp.Interval),
		Open:      convert.QuotationToDec(resp.Open),
		High:      convert.QuotationToDec(resp.High),
		Low:       convert.QuotationToDec(resp.Low),
		Close:     convert.QuotationToDec(resp.Close),
		Volume:    resp.Volume,
		TimeStart: resp.Time.AsTime(),
		TimeEnd:   resp.LastTradeTs.AsTime(),
	}
}
