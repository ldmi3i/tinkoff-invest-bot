package tapi

import (
	"github.com/shopspring/decimal"
	"invest-robot/convert"
	investapi "invest-robot/tapigen"
)

type SecurityTradingStatus int
type RealExchange int

const (
	SECURITY_TRADING_STATUS_UNSPECIFIED               SecurityTradingStatus = iota //Undefined status
	SECURITY_TRADING_STATUS_NOT_AVAILABLE_FOR_TRADING                              //Not available for trading
	SECURITY_TRADING_STATUS_OPENING_PERIOD
	SECURITY_TRADING_STATUS_CLOSING_PERIOD
	SECURITY_TRADING_STATUS_BREAK_IN_TRADING
	SECURITY_TRADING_STATUS_NORMAL_TRADING
	SECURITY_TRADING_STATUS_CLOSING_AUCTION
	SECURITY_TRADING_STATUS_DARK_POOL_AUCTION
	SECURITY_TRADING_STATUS_DISCRETE_AUCTION
	SECURITY_TRADING_STATUS_OPENING_AUCTION_PERIOD
	SECURITY_TRADING_STATUS_TRADING_AT_CLOSING_AUCTION_PRICE
	SECURITY_TRADING_STATUS_SESSION_ASSIGNED
	SECURITY_TRADING_STATUS_SESSION_CLOSE
	SECURITY_TRADING_STATUS_SESSION_OPEN
	SECURITY_TRADING_STATUS_DEALER_NORMAL_TRADING
	SECURITY_TRADING_STATUS_DEALER_BREAK_IN_TRADING
	SECURITY_TRADING_STATUS_DEALER_NOT_AVAILABLE_FOR_TRADING
)

const (
	REAL_EXCHANGE_UNSPECIFIED RealExchange = iota
	REAL_EXCHANGE_MOEX                     //Moscow exchange
	REAL_EXCHANGE_RTS                      //Saint-Petersburg exchange
	REAL_EXCHANGE_OTC                      //External instrument
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
	return ir.TradingStatus == SECURITY_TRADING_STATUS_NORMAL_TRADING ||
		ir.TradingStatus == SECURITY_TRADING_STATUS_DEALER_NORMAL_TRADING

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
