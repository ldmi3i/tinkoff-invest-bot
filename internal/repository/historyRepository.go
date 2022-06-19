package repository

import (
	"fmt"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/collections"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"gorm.io/gorm"
)

//HistoryRepository provides methods to operate domain.History database data
//go:generate mockgen -source=historyRepository.go -destination=../mocks/repository/mockHistoryRepository.go -package=repository
type HistoryRepository interface {
	ClearAndSaveAll(history []entity.History) error
	FindAll() ([]entity.History, error)
	FindAllByFigis(figis []string) ([]entity.History, error)
}

type PgHistoryRepository struct {
	db                  *gorm.DB
	findAllByFigisCache collections.SyncMap[string, []entity.History]
}

func (h PgHistoryRepository) ClearAndSaveAll(history []entity.History) error {
	if err := h.db.Exec("DELETE FROM history").Error; err != nil {
		return err
	}
	if err := h.db.Create(history).Error; err != nil {
		return err
	}
	h.findAllByFigisCache.Clear()
	return nil
}

func (h PgHistoryRepository) FindAll() ([]entity.History, error) {
	var hist []entity.History
	if err := h.db.Order("date").Find(&hist).Error; err != nil {
		return nil, err
	}
	return hist, nil
}

func (h PgHistoryRepository) FindAllByFigis(figis []string) ([]entity.History, error) {
	strKey := fmt.Sprint(figis)
	hist, ok := h.findAllByFigisCache.Get(strKey)
	if ok {
		return hist, nil
	}
	if err := h.db.Where("figi in ?", figis).Order("time").Find(&hist).Error; err != nil {
		return nil, err
	}
	h.findAllByFigisCache.Put(strKey, hist)
	return hist, nil
}

func NewHistoryRepository(db *gorm.DB) HistoryRepository {
	return PgHistoryRepository{db, collections.NewSyncMap[string, []entity.History]()}
}
