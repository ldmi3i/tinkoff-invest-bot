package dto

import "github.com/shopspring/decimal"

type CreateAlgorithmRequest struct {
	Figis    []string
	Strategy string
	Amount   decimal.Decimal
	Params   map[string]string
}
