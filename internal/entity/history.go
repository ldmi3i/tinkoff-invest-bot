package entity

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/convert"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"github.com/shopspring/decimal"
	"time"
)

//History represents one row of downloaded history. I.e. candle parameters.
//Constructs from Candle response from API
type History struct {
	ID    uint `gorm:"primaryKey"`
	Figi  string
	Open  decimal.Decimal `gorm:"type:numeric"` //Open price
	Low   decimal.Decimal `gorm:"type:numeric"` //Lowest price
	High  decimal.Decimal `gorm:"type:numeric"` //Highest price
	Close decimal.Decimal `gorm:"type:numeric"` //Close price
	Time  time.Time       //Timestamp of history record
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
