package tapi

import (
	investapi "invest-robot/tapigen"
	"time"
)

type CancelOrderResponse struct {
	Time time.Time
}

func CancelOrderResponseToDto(resp *investapi.CancelOrderResponse) *CancelOrderResponse {
	return &CancelOrderResponse{
		Time: resp.Time.AsTime(),
	}
}
