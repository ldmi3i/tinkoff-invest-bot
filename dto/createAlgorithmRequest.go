package dto

import "github.com/shopspring/decimal"

type CreateAlgorithmRequest struct {
	AccountId string            `json:"accountId"`
	Figis     []string          `json:"figis"`
	Strategy  string            `json:"strategy"`
	Limits    []MoneyValue      `json:"limits"`
	Params    map[string]string `json:"params"`
}

type MoneyValue struct {
	Currency string          `json:"currency"`
	Value    decimal.Decimal `json:"value"`
}
