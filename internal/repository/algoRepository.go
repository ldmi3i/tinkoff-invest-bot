package repository

import (
	"gorm.io/gorm"
	"invest-robot/internal/domain"
	"invest-robot/internal/errors"
)

//AlgoRepository provides methods to operate domain.Algorithm database data
type AlgoRepository interface {
	Save(algo *domain.Algorithm) error
	SetActiveStatus(id uint, isActive bool) error
}

type PgAlgoRepository struct {
	db *gorm.DB
}

func (ar *PgAlgoRepository) SetActiveStatus(id uint, isActive bool) error {
	sql := "update algorithms set is_active = ? where id = ?"
	if err := ar.db.Exec(sql, isActive, id).Error; err != nil {
		return err
	}
	return nil
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
