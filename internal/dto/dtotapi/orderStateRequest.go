package dtotapi

import "github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"

type OrderStateRequest struct {
	AccountId string
	OrderId   string
}

func (r *OrderStateRequest) ToTinApi() *investapi.GetOrderStateRequest {
	return &investapi.GetOrderStateRequest{
		AccountId: r.AccountId,
		OrderId:   r.OrderId,
	}
}
