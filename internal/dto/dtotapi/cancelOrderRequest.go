package dtotapi

import "github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"

type CancelOrderRequest struct {
	AccountId string
	OrderId   string
}

func (cor *CancelOrderRequest) ToTinApi() *investapi.CancelOrderRequest {
	return &investapi.CancelOrderRequest{
		AccountId: cor.AccountId,
		OrderId:   cor.OrderId,
	}
}
