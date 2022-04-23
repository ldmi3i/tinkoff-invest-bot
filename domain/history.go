package domain

import (
	"github.com/shopspring/decimal"
	investapi "invest-robot/tapigen"
	"time"
)

type History struct {
	ID    uint `gorm:"primaryKey"`
	Open  decimal.Decimal
	Low   decimal.Decimal
	High  decimal.Decimal
	Close decimal.Decimal
	Date  time.Time
}

func FromCandle(c *investapi.Candle) History {
	return History{
		ID:    0,
		Open:  decimal.New(c.Open.Units, c.Open.Nano),
		Low:   decimal.New(c.Low.Units, c.Low.Nano),
		High:  decimal.New(c.High.Units, c.High.Nano),
		Close: decimal.New(c.Close.Units, c.Close.Nano),
		Date:  c.Time.AsTime(),
	}
}

func FromHistoricCandle(c *investapi.HistoricCandle) History {
	return History{
		ID:    0,
		Open:  decimal.New(c.Open.Units, c.Open.Nano),
		Low:   decimal.New(c.Low.Units, c.Low.Nano),
		High:  decimal.New(c.High.Units, c.High.Nano),
		Close: decimal.New(c.Close.Units, c.Close.Nano),
		Date:  c.Time.AsTime(),
	}
}
