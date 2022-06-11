package dtotapi

import investapi "invest-robot/internal/tapigen"

type PositionsRequest struct {
	AccountId string
}

func (req *PositionsRequest) ToTinApi() *investapi.PositionsRequest {
	return &investapi.PositionsRequest{AccountId: req.AccountId}
}
