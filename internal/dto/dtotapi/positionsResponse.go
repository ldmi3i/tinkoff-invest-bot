package dtotapi

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
)

type PositionsResponse struct {
	Money                   []*MoneyValue
	Blocked                 []*MoneyValue
	Securities              []*PositionsSecurity
	LimitsLoadingInProgress bool
	Futures                 []*PositionsFuture
}

func (ps *PositionsResponse) GetMoney(currency string) *MoneyValue {
	for _, mn := range ps.Money {
		if mn.Currency == currency {
			return mn
		}
	}
	return nil
}

func (ps *PositionsResponse) GetInstrument(figi string) *PositionsSecurity {
	for _, security := range ps.Securities {
		if security.Figi == figi {
			return security
		}
	}
	return nil
}

type PositionsSecurity struct {
	Figi    string
	Blocked int64
	Balance int64
}

func positionsSecurityToDto(resp *investapi.PositionsSecurities) *PositionsSecurity {
	return &PositionsSecurity{
		Figi:    resp.Figi,
		Blocked: resp.Blocked,
		Balance: resp.Balance,
	}
}

type PositionsFuture struct {
	Figi    string
	Blocked int64
	Balance int64
}

func positionsFutureToDto(resp *investapi.PositionsFutures) *PositionsFuture {
	return &PositionsFuture{
		Figi:    resp.Figi,
		Blocked: resp.Blocked,
		Balance: resp.Balance,
	}
}

func PositionsResponseToDto(resp *investapi.PositionsResponse) *PositionsResponse {
	money := make([]*MoneyValue, 0, len(resp.Money))
	for _, mn := range resp.Money {
		money = append(money, MoneyValueToDto(mn))
	}
	blocked := make([]*MoneyValue, 0, len(resp.Blocked))
	for _, bl := range resp.Blocked {
		blocked = append(blocked, MoneyValueToDto(bl))
	}
	securities := make([]*PositionsSecurity, 0, len(resp.Securities))
	for _, security := range resp.Securities {
		securities = append(securities, positionsSecurityToDto(security))
	}
	futures := make([]*PositionsFuture, 0, len(resp.Futures))
	for _, future := range resp.Futures {
		futures = append(futures, positionsFutureToDto(future))
	}
	return &PositionsResponse{
		Money:                   money,
		Blocked:                 blocked,
		Securities:              securities,
		LimitsLoadingInProgress: resp.LimitsLoadingInProgress,
		Futures:                 futures,
	}
}
