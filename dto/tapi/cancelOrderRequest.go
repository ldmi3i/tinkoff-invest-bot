package tapi

import investapi "invest-robot/tapigen"

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
