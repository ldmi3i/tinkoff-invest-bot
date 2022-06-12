package dtotapi

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
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
