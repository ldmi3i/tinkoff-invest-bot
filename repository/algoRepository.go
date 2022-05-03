package repository

import (
	"gorm.io/gorm"
	"invest-robot/helper"
)

type AlgoRepository interface {
}

type PgAlgoRepository struct {
	db *gorm.DB
}

func NewAlgoRepository() AlgoRepository {
	return PgHistoryRepository{helper.GetDB()}
}
