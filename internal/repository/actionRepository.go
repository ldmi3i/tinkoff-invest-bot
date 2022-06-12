package repository

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/domain"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/errors"
	"gorm.io/gorm"
	"log"
)

//ActionRepository provides methods to operate actions database data
type ActionRepository interface {
	Save(action *domain.Action) error
	UpdateStatusWithMsg(id uint, status domain.ActionStatus, msg string) error
}

type PgActionRepository struct {
	db *gorm.DB
}

func (rep *PgActionRepository) Save(action *domain.Action) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Save method failed and recovered, info: %s", r)
			err = errors.ConvertToError(r)
		}
	}()
	return rep.db.Save(action).Error
}

func (rep *PgActionRepository) UpdateStatusWithMsg(id uint, status domain.ActionStatus, msg string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("UpdateStatusWithMsg method failed and recovered, info: %s", r)
			err = errors.ConvertToError(r)
		}
	}()
	return rep.db.Model(&domain.Action{}).Where("id = ?", id).Updates(domain.Action{Status: status, Info: msg}).Error
}

func NewActionRepository(db *gorm.DB) ActionRepository {
	return &PgActionRepository{db: db}
}
