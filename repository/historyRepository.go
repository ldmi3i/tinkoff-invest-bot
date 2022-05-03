package repository

import (
	"gorm.io/gorm"
	"invest-robot/domain"
)

type HistoryRepository interface {
	SaveAll(history []domain.History) error
	FindAll() ([]domain.History, error)
	FindAllByFigis(figis []string) ([]domain.History, error)
}

type PgHistoryRepository struct {
	db *gorm.DB
}

func (h PgHistoryRepository) SaveAll(history []domain.History) error {
	h.db.Exec("DELETE FROM history")
	h.db.Create(history)
	return nil
}

func (h PgHistoryRepository) FindAll() ([]domain.History, error) {
	var hist []domain.History
	h.db.Order("date").Find(&hist)
	return hist, nil
}

func (h PgHistoryRepository) FindAllByFigis(figis []string) ([]domain.History, error) {
	var hist []domain.History
	h.db.Where("figi in ?", figis).Order("time").Find(&hist)
	return hist, nil
}

func NewHistoryRepository(db *gorm.DB) HistoryRepository {
	return PgHistoryRepository{db}
}
