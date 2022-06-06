package repository

import (
	"fmt"
	"gorm.io/gorm"
	"invest-robot/collections"
	"invest-robot/domain"
)

//HistoryRepository provides methods to operate domain.History database data
//go:generate mockgen -source=historyRepository.go -destination=../mocks/repository/mockHistoryRepository.go -package=repository
type HistoryRepository interface {
	ClearAndSaveAll(history []domain.History) error
	FindAll() ([]domain.History, error)
	FindAllByFigis(figis []string) ([]domain.History, error)
}

type PgHistoryRepository struct {
	db                  *gorm.DB
	findAllByFigisCache collections.SyncMap[string, []domain.History]
}

func (h PgHistoryRepository) ClearAndSaveAll(history []domain.History) error {
	h.db.Exec("DELETE FROM history")
	h.db.Create(history)
	h.findAllByFigisCache.Clear()
	return nil
}

func (h PgHistoryRepository) FindAll() ([]domain.History, error) {
	var hist []domain.History
	h.db.Order("date").Find(&hist)
	return hist, nil
}

func (h PgHistoryRepository) FindAllByFigis(figis []string) ([]domain.History, error) {
	strKey := fmt.Sprint(figis)
	hist, ok := h.findAllByFigisCache.Get(strKey)
	if ok {
		return hist, nil
	}
	h.db.Where("figi in ?", figis).Order("time").Find(&hist)
	h.findAllByFigisCache.Put(strKey, hist)
	return hist, nil
}

func NewHistoryRepository(db *gorm.DB) HistoryRepository {
	return PgHistoryRepository{db, collections.NewSyncMap[string, []domain.History]()}
}
