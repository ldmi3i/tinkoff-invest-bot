package dtotapi

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/convert"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"github.com/shopspring/decimal"
	"time"
)

type LastPricesResponse struct {
	LastPrices []*LastPrice
}

func (lpr *LastPricesResponse) GetByFigi(figi string) *LastPrice {
	for _, lp := range lpr.LastPrices {
		if lp.Figi == figi {
			return lp
		}
	}
	return nil
}

type LastPrice struct {
	Figi  string
	Price decimal.Decimal
	Time  time.Time
}

func LastPricesResponseToDto(resp *investapi.GetLastPricesResponse) *LastPricesResponse {
	lastPrices := make([]*LastPrice, 0, len(resp.LastPrices))
	for _, lp := range resp.LastPrices {
		lastPrices = append(lastPrices, lastPriceToDto(lp))
	}
	return &LastPricesResponse{LastPrices: lastPrices}
}

func lastPriceToDto(lp *investapi.LastPrice) *LastPrice {
	return &LastPrice{
		Figi:  lp.Figi,
		Price: convert.QuotationToDec(lp.Price),
		Time:  lp.Time.AsTime(),
	}
}
