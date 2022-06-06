package service

import (
	"go.uber.org/zap"
	"invest-robot/dto"
	"invest-robot/repository"
)

//go:generate mockgen -source=statService.go -destination=../mocks/service/statService.go -package=service
type StatService interface {
	//GetAlgorithmStat returns statistics for the algorithm id
	GetAlgorithmStat(req *dto.StatAlgoRequest) (*dto.StatAlgoResponse, error)
}

type StatServiceImpl struct {
	statRep repository.StatRepository
	logger  *zap.SugaredLogger
}

func NewStatService(statRep repository.StatRepository, logger *zap.SugaredLogger) StatService {
	return &StatServiceImpl{
		statRep: statRep,
		logger:  logger,
	}
}

func (ss *StatServiceImpl) GetAlgorithmStat(req *dto.StatAlgoRequest) (*dto.StatAlgoResponse, error) {
	return ss.statRep.GetAlgorithmStat(req)
}
