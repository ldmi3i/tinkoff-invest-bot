package repository

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/errors"
	"gorm.io/gorm"
	"log"
)

//ActionRepository provides methods to operate actions database data
type ActionRepository interface {
	Save(action *entity.Action) error
	UpdateStatusWithMsg(id uint, status entity.ActionStatus, msg string) error
}

type PgActionRepository struct {
	db *gorm.DB
}

func (rep *PgActionRepository) Save(action *entity.Action) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Save method failed and recovered, info: %s", r)
			err = errors.ConvertToError(r)
		}
	}()
	return rep.db.Save(action).Error
}

func (rep *PgActionRepository) UpdateStatusWithMsg(id uint, status entity.ActionStatus, msg string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("UpdateStatusWithMsg method failed and recovered, info: %s", r)
			err = errors.ConvertToError(r)
		}
	}()
	return rep.db.Model(&entity.Action{}).Where("id = ?", id).Updates(entity.Action{Status: status, Info: msg}).Error
}

func NewActionRepository(db *gorm.DB) ActionRepository {
	return &PgActionRepository{db: db}
}
