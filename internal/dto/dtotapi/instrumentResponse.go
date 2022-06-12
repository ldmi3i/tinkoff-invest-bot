package dtotapi

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/convert"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"github.com/shopspring/decimal"
)

type SecurityTradingStatus int
type RealExchange int

const (
	SecurityTradingStatusUnspecified            SecurityTradingStatus = iota //Undefined status
	SecurityTradingStatusNotAvailableForTrading                              //Not available for trading
	SecurityTradingStatusOpeningPeriod
	SecurityTradingStatusClosingPeriod
	SecurityTradingStatusBreakInTrading
	SecurityTradingStatusNormalTrading
	SecurityTradingStatusClosingAuction
	SecurityTradingStatusDarkPoolAuction
	SecurityTradingStatusDiscreteAuction
	SecurityTradingStatusOpeningAuctionPeriod
	SecurityTradingStatusTradingAtClosingAuctionPrice
	SecurityTradingStatusSessionAssigned
	SecurityTradingStatusSessionClose
	SecurityTradingStatusSessionOpen
	SecurityTradingStatusDealerNormalTrading
	SecurityTradingStatusDealerBreakInTrading
	SecurityTradingStatusDealerNotAvailableForTrading
)

const (
	RealExchangeUnspecified RealExchange = iota
	RealExchangeMoex                     //Moscow exchange
	RealExchangeRts                      //Saint-Petersburg exchange
	RealExchangeOtc                      //External instrument
)

type InstrumentResponse struct {
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

	CountryOfRisk     string
	CountryOfRiskName string
	InstrumentType    string

	TradingStatus         SecurityTradingStatus
	OtcFlag               bool
	BuyAvailableFlag      bool
	SellAvailableFlag     bool
	MinPriceIncrement     decimal.Decimal
	ApiTradeAvailableFlag bool

	Uid          string
	RealExchange RealExchange
}

func (ir *InstrumentResponse) IsTradingAvailable() bool {
	return ir.TradingStatus == SecurityTradingStatusNormalTrading ||
		ir.TradingStatus == SecurityTradingStatusDealerNormalTrading

}

func InstrumentResponseToDto(res *investapi.InstrumentResponse) *InstrumentResponse {
	instr := res.Instrument
	return &InstrumentResponse{
		Figi:                  instr.Figi,
		Ticker:                instr.Ticker,
		ClassCode:             instr.ClassCode,
		Isin:                  instr.Isin,
		Lot:                   int64(instr.Lot),
		Currency:              instr.Currency,
		KLong:                 convert.QuotationToDec(instr.Klong),
		KShort:                convert.QuotationToDec(instr.Kshort),
		DLong:                 convert.QuotationToDec(instr.Dlong),
		DShort:                convert.QuotationToDec(instr.Dshort),
		DLongMin:              convert.QuotationToDec(instr.DlongMin),
		DShortMin:             convert.QuotationToDec(instr.DshortMin),
		ShortEnabledFlag:      instr.ShortEnabledFlag,
		Name:                  instr.Name,
		Exchange:              instr.Exchange,
		CountryOfRisk:         instr.CountryOfRisk,
		CountryOfRiskName:     instr.CountryOfRiskName,
		InstrumentType:        instr.InstrumentType,
		TradingStatus:         SecurityTradingStatus(instr.TradingStatus),
		OtcFlag:               instr.OtcFlag,
		BuyAvailableFlag:      instr.BuyAvailableFlag,
		SellAvailableFlag:     instr.SellAvailableFlag,
		MinPriceIncrement:     convert.QuotationToDec(instr.MinPriceIncrement),
		ApiTradeAvailableFlag: instr.ApiTradeAvailableFlag,
		Uid:                   instr.Uid,
		RealExchange:          RealExchange(instr.RealExchange),
	}
}
