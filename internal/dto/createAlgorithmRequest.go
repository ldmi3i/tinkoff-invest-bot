package dto

import (
	"github.com/shopspring/decimal"
)

//CreateAlgorithmRequest request to create new trade algorithm or history algorithm config
type CreateAlgorithmRequest struct {
	AccountId string            `json:"accountId"`
	Figis     []string          `json:"figis"`
	Strategy  string            `json:"strategy"`
	Limits    []MoneyValue      `json:"limits"` //Algorithm limits on using money
	Params    map[string]string `json:"params"`
	InstrInit *InstrumentsInfo  `json:"instrInit"` //Optional - to specify initial instrument available
}

type MoneyValue struct {
	Currency string          `json:"currency"`
	Value    decimal.Decimal `json:"value"`
}

type InstrumentValue struct {
	Figi   string          `json:"figi"`
	Amount int64           `json:"amount"`
	Price  decimal.Decimal `json:"price"`
}
