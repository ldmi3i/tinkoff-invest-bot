package repository

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/errors"
	"gorm.io/gorm"
)

//AlgoRepository provides methods to operate domain.Algorithm database data
type AlgoRepository interface {
	Save(algo *entity.Algorithm) error
	SetActiveStatus(id uint, isActive bool) error
}

type PgAlgoRepository struct {
	db *gorm.DB
}

func (ar *PgAlgoRepository) SetActiveStatus(id uint, isActive bool) error {
	sql := "update algorithms set is_active = ? where id = ?"
	return ar.db.Exec(sql, isActive, id).Error
}

func (ar *PgAlgoRepository) Save(algo *entity.Algorithm) (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = errors.ConvertToError(rec)
		}
	}()
	return ar.db.Save(algo).Error
}

func NewAlgoRepository(db *gorm.DB) AlgoRepository {
	return &PgAlgoRepository{db: db}
}
