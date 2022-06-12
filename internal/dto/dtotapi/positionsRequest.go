package dtotapi

import "github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"

type PositionsRequest struct {
	AccountId string
}

func (req *PositionsRequest) ToTinApi() *investapi.PositionsRequest {
	return &investapi.PositionsRequest{AccountId: req.AccountId}
}
