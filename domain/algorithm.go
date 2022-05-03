package domain

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"invest-robot/dto"
)

type Algorithm struct {
	gorm.Model
	Strategy   string
	Figis      []string          `gorm:"type:text[]"`
	Currencies []string          `gorm:"type:text[]"`
	Limits     []decimal.Decimal `gorm:"type:numeric[]"`
	Params     []Param
	CtxParams  []CtxParam
	Actions    []Action
	IsActive   bool
}

type Param struct {
	AlgorithmID uint
	Key         string
	Value       string
}

type CtxParam struct {
	AlgorithmID uint
	Key         string
	Value       string
}

func ParamsToMap(params []Param) map[string]string {
	res := make(map[string]string, len(params))
	for _, param := range params {
		res[param.Key] = param.Value
	}
	return res
}

func AlgorithmFromDto(req dto.CreateAlgorithmRequest) Algorithm {
	params := make([]Param, 0, len(req.Params))
	for key, val := range req.Params {
		params = append(params, Param{Key: key, Value: val})
	}
	return Algorithm{
		Strategy: req.Strategy,
		Figis:    req.Figis,
		Params:   params,
		IsActive: true,
	}
}
