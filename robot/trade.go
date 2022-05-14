package robot

import (
	"invest-robot/dto"
)

type TradeAPI interface {
	Trade(req *dto.CreateAlgorithmRequest) (*dto.TradeStartResponse, error)
}
