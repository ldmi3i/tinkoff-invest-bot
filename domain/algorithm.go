package domain

import (
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"invest-robot/dto"
)

type Algorithm struct {
	gorm.Model
	Strategy    string
	AccountId   string
	Figis       pq.StringArray `gorm:"type:text[]"`
	MoneyLimits []*MoneyLimit
	Params      []*Param
	CtxParams   []*CtxParam
	Actions     []*Action
	IsActive    bool
}

type Param struct {
	ID          uint `gorm:"primaryKey"`
	AlgorithmID uint
	Key         string
	Value       string
}

type CtxParam struct {
	ID          uint `gorm:"primaryKey"`
	AlgorithmID uint
	Key         string
	Value       string
}

type MoneyLimit struct {
	ID          uint `gorm:"primaryKey"`
	AlgorithmID uint
	Currency    string
	Amount      decimal.Decimal
}

func ParamsToMap(params []*Param) map[string]string {
	res := make(map[string]string, len(params))
	for _, param := range params {
		res[param.Key] = param.Value
	}
	return res
}

func (alg *Algorithm) CopyNoParam() *Algorithm {
	return &Algorithm{
		Strategy:    alg.Strategy,
		Figis:       alg.Figis,
		MoneyLimits: alg.MoneyLimits, //Copy is ok - not modifiable
		AccountId:   alg.AccountId,
		CtxParams:   alg.CtxParams,
		IsActive:    alg.IsActive,
	}
}

func AlgorithmFromDto(req *dto.CreateAlgorithmRequest) *Algorithm {
	params := make([]*Param, 0, len(req.Params))
	for key, val := range req.Params {
		params = append(params, &Param{Key: key, Value: val})
	}
	limits := make([]*MoneyLimit, 0, len(req.Limits))
	for _, lim := range req.Limits {
		mLim := MoneyLimit{
			Currency: lim.Currency,
			Amount:   lim.Value,
		}
		limits = append(limits, &mLim)
	}
	return &Algorithm{
		Strategy:    req.Strategy,
		Figis:       req.Figis,
		MoneyLimits: limits,
		Params:      params,
		AccountId:   req.AccountId,
		IsActive:    true,
	}
}
