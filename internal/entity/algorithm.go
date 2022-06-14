package entity

import (
	"encoding/json"
	"fmt"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

//Algorithm represents full algorithm configuration
type Algorithm struct {
	gorm.Model
	Strategy    string         //Name of strategy
	AccountId   string         //Account identity
	Figis       pq.StringArray `gorm:"type:text[]"` //List of figis to operate
	MoneyLimits []*MoneyLimit  //OneToMany List of money limits to operate
	Params      []*Param       //OneToMany List of algorithm parameters
	CtxParams   []*CtxParam    //OneToMany Context algorithm parameters required to save/restore state TODO not implemeted!
	Actions     []*Action      //OneToMany List of actions made by algorithm
	IsActive    bool           //Algorithm activity state, if running - true, else - false
}

func (alg *Algorithm) ToDto() *dto.AlgorithmResponse {
	limits := make([]*dto.MoneyValue, 0, len(alg.MoneyLimits))
	for _, lim := range alg.MoneyLimits {
		limits = append(limits, lim.ToDto())
	}
	algResp := dto.AlgorithmResponse{
		AlgorithmID: alg.ID,
		Strategy:    alg.Strategy,
		AccountId:   alg.AccountId,
		Figis:       alg.Figis,
		MoneyLimits: limits,
		Params:      ParamsToMap(alg.Params),
		IsActive:    alg.IsActive,
		CreatedAt:   alg.CreatedAt,
		UpdatedAt:   alg.UpdatedAt,
	}

	if instrInfoStr, ok := alg.GetCtxParam(dto.InstrAmountField); ok {
		var instrInfo dto.InstrumentsInfo
		if err := json.Unmarshal([]byte(instrInfoStr.Value), &instrInfo); err == nil {
			algResp.InstrAvail = &instrInfo
		}
	}
	return &algResp
}

//Param represents algorithm parametrization by key/value map
type Param struct {
	ID          uint `gorm:"primaryKey"`
	AlgorithmID uint
	Key         string
	Value       string
}

//CtxParam represents algorithm current calculation state with key/value map
type CtxParam struct {
	ID          uint `gorm:"primaryKey"`
	AlgorithmID uint
	Key         string
	Value       string
}

func (ctp *CtxParam) String() string {
	return fmt.Sprintf("ID: %d, AlgorithmID: %d, Key: %s, Value: %s", ctp.ID, ctp.AlgorithmID, ctp.Key, ctp.Value)
}

//MoneyLimit represents limits on the use of money
type MoneyLimit struct {
	ID          uint `gorm:"primaryKey"`
	AlgorithmID uint
	Currency    string
	Amount      decimal.Decimal
}

func (ml *MoneyLimit) ToDto() *dto.MoneyValue {
	return &dto.MoneyValue{
		Currency: ml.Currency,
		Value:    ml.Amount,
	}
}

func ParamsToMap(params []*Param) map[string]string {
	res := make(map[string]string, len(params))
	for _, param := range params {
		res[param.Key] = param.Value
	}
	return res
}

func ContextToMap(params []*CtxParam) map[string]string {
	res := make(map[string]string, len(params))
	for _, param := range params {
		res[param.Key] = param.Value
	}
	return res
}

//CopyNoParam is utility method for parameter variation.
//Copies algorithm without parameters
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

//GetCtxParam returns CtxParam by param name and flag is requested parameter exits
func (alg *Algorithm) GetCtxParam(paramName string) (*CtxParam, bool) {
	for _, param := range alg.CtxParams {
		if param.Key == paramName {
			return param, true
		}
	}
	return nil, false
}

func AlgorithmFromDto(req *dto.CreateAlgorithmRequest) *Algorithm {
	params := make([]*Param, 0, len(req.Params))
	for key, val := range req.Params {
		params = append(params, &Param{Key: key, Value: val})
	}

	//Pre-populate context with serialized initial amount of available instruments
	//Set instrument amount to context because it's relates to algorithm state, not configuration
	ctxParam := make([]*CtxParam, 0)
	if val, err := json.Marshal(req.InstrInit); err == nil {
		param := CtxParam{
			Key:   dto.InstrAmountField,
			Value: string(val),
		}
		ctxParam = append(ctxParam, &param)
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
		CtxParams:   ctxParam,
		IsActive:    true,
	}
}
