package dtotapi

import investapi "invest-robot/tapigen"

type GetOrderStateRequest struct {
	AccountId string
	OrderId   string
}

func (r *GetOrderStateRequest) ToTinApi() *investapi.GetOrderStateRequest {
	return &investapi.GetOrderStateRequest{
		AccountId: r.AccountId,
		OrderId:   r.OrderId,
	}
}
