package repository

import (
	"gorm.io/gorm"
	"invest-robot/domain"
	"invest-robot/errors"
)

type AlgoRepository interface {
	Save(algo *domain.Algorithm) error
}

type PgAlgoRepository struct {
	db *gorm.DB
}

func (ar *PgAlgoRepository) Save(algo *domain.Algorithm) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = errors.ConvertToError(rec)
		}
	}()
	ar.db.Save(algo)
	return nil
}

func NewAlgoRepository(db *gorm.DB) AlgoRepository {
	return &PgAlgoRepository{db: db}
}
