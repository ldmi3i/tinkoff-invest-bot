package dtotapi

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/convert"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"github.com/shopspring/decimal"
	"time"
)

type SharesResponse struct {
	Instruments []*ShareResponse
}

type ShareResponse struct {
	Figi      string
	Ticker    string
	ClassCode string
	Isin      string
	Lot       int64
	Currency  string

	KLong            decimal.Decimal
	KShort           decimal.Decimal
	DLong            decimal.Decimal
	DShort           decimal.Decimal
	DLongMin         decimal.Decimal
	DShortMin        decimal.Decimal
	ShortEnabledFlag bool
	Name             string
	Exchange         string

	IpoDate   time.Time
	IssueSize int64

	CountryOfRisk     string
	CountryOfRiskName string
	Sector            string
	IssueSizePlan     int64
	Nominal           *MoneyValue

	TradingStatus         SecurityTradingStatus
	OtcFlag               bool
	BuyAvailableFlag      bool
	SellAvailableFlag     bool
	DivYieldFlag          bool
	ShareType             int
	MinPriceIncrement     decimal.Decimal
	ApiTradeAvailableFlag bool

	Uid          string
	RealExchange RealExchange
}

func SharesResponseToDto(res *investapi.SharesResponse) *SharesResponse {
	sharesRes := make([]*ShareResponse, 0, len(res.Instruments))
	for _, share := range res.Instruments {
		sharesRes = append(sharesRes, shareResponseToDto(share))
	}

	return &SharesResponse{
		Instruments: sharesRes,
	}
}

func shareResponseToDto(share *investapi.Share) *ShareResponse {
	return &ShareResponse{
		Figi:      share.Figi,
		Ticker:    share.Ticker,
		ClassCode: share.ClassCode,
		Isin:      share.Isin,
		Lot:       int64(share.Lot),
		Currency:  share.Currency,

		KLong:            convert.QuotationToDec(share.Klong),
		KShort:           convert.QuotationToDec(share.Kshort),
		DLong:            convert.QuotationToDec(share.Dlong),
		DShort:           convert.QuotationToDec(share.Dshort),
		DLongMin:         convert.QuotationToDec(share.DlongMin),
		DShortMin:        convert.QuotationToDec(share.DshortMin),
		ShortEnabledFlag: share.ShortEnabledFlag,
		Name:             share.Name,
		Exchange:         share.Exchange,

		IpoDate:   share.IpoDate.AsTime(),
		IssueSize: share.IssueSize,

		CountryOfRisk:     share.CountryOfRisk,
		CountryOfRiskName: share.CountryOfRiskName,
		Sector:            share.Sector,
		IssueSizePlan:     share.IssueSizePlan,
		Nominal:           MoneyValueToDto(share.Nominal),

		TradingStatus:         SecurityTradingStatus(share.TradingStatus),
		OtcFlag:               share.OtcFlag,
		BuyAvailableFlag:      share.BuyAvailableFlag,
		SellAvailableFlag:     share.SellAvailableFlag,
		DivYieldFlag:          share.DivYieldFlag,
		ShareType:             int(share.ShareType),
		MinPriceIncrement:     convert.QuotationToDec(share.MinPriceIncrement),
		ApiTradeAvailableFlag: share.ApiTradeAvailableFlag,

		Uid:          share.Uid,
		RealExchange: RealExchange(share.RealExchange),
	}
}
