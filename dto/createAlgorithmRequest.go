package dto

import "github.com/shopspring/decimal"

type CreateAlgorithmRequest struct {
	AccountId  string            `json:"accountId"`
	Figis      []string          `json:"figis"`
	Strategy   string            `json:"strategy"`
	Currencies []string          `json:"currencies"`
	Limits     []decimal.Decimal `json:"limits"`
	Params     map[string]string `json:"params"`
}
