package dto

import (
	"time"
)

type AlgorithmsResponse struct {
	Algorithms []*AlgorithmResponse `json:"algorithms"`
}

type AlgorithmResponse struct {
	AlgorithmID uint
	Strategy    string            `json:"strategy"`
	AccountId   string            `json:"accountId"`
	Figis       []string          `json:"figis"`
	MoneyLimits []*MoneyValue     `json:"moneyLimits"`
	Params      map[string]string `json:"params"`
	IsActive    bool              `json:"isActive"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}
