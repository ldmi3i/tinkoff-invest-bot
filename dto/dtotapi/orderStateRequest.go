package dtotapi

import investapi "invest-robot/tapigen"

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
