package domain

import (
	"github.com/shopspring/decimal"
	"invest-robot/convert"
	investapi "invest-robot/tapigen"
	"time"
)

type History struct {
	ID    uint `gorm:"primaryKey"`
	Figi  string
	Open  decimal.Decimal `gorm:"type:numeric"`
	Low   decimal.Decimal `gorm:"type:numeric"`
	High  decimal.Decimal `gorm:"type:numeric"`
	Close decimal.Decimal `gorm:"type:numeric"`
	Time  time.Time
}

func FromCandle(c *investapi.Candle) History {
	return History{
		ID:    0,
		Open:  convert.QuotationToDec(c.Open),
		Low:   convert.QuotationToDec(c.Low),
		High:  convert.QuotationToDec(c.High),
		Close: convert.QuotationToDec(c.Close),
		Time:  c.Time.AsTime(),
	}
}

func FromHistoricCandle(c *investapi.HistoricCandle) History {
	return History{
		ID:    0,
		Open:  convert.QuotationToDec(c.Open),
		Low:   convert.QuotationToDec(c.Low),
		High:  convert.QuotationToDec(c.High),
		Close: convert.QuotationToDec(c.Close),
		Time:  c.Time.AsTime(),
	}
}

func (History) TableName() string {
	return "history"
}
