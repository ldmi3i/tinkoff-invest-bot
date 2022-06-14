package avr

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/errors"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/strategy/stmodel"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"regexp"
)

const (
	separator string = ":"
)

type ParamSplitterImpl struct {
	logger *zap.SugaredLogger
}

func (p *ParamSplitterImpl) ParseAndSplit(param map[string]string) ([]map[string]string, error) {
	shortStr, sdExt := param[ShortDur]
	longStr, ldExt := param[LongDur]
	if !sdExt || !ldExt {
		p.logger.Error("Short duration range and long duration range must be defined")
		return nil, errors.NewUnexpectedError("One of short or long ranges not defined")
	}
	sepRgx := regexp.MustCompile("\\s*" + separator + "\\s*")
	shortLimits := sepRgx.Split(shortStr, -1)
	if len(shortLimits) > 3 {
		p.logger.Error("Short range has wrong format, split results in: ", shortLimits)
		return nil, errors.NewUnexpectedError("Short range has wrong format!")
	}
	longLimits := sepRgx.Split(longStr, -1)
	if len(longLimits) > 3 {
		p.logger.Error("Long range has wrong format, split results in: ", longLimits)
		return nil, errors.NewUnexpectedError("Long range has wrong format!")
	}
	shortRange, err := p.convertToRange(shortLimits)
	if err != nil {
		p.logger.Error("Error while converting short duration expression to range: ", shortLimits)
		return nil, err
	}
	longRange, err := p.convertToRange(longLimits)
	if err != nil {
		p.logger.Error("Error while converting long duration expression to range: ", longLimits)
		return nil, err
	}
	//Params to copy to all algorithm copies
	constParam := make(map[string]string)
	for key, val := range param {
		if key != ShortDur && key != LongDur {
			constParam[key] = val
		}
	}
	resultMap := make([]map[string]string, 0)
	//Iterate over long range and make all pairs with lowers short ranges
	for _, longAvr := range longRange {
		for _, shortAvr := range shortRange {
			if shortAvr.GreaterThanOrEqual(longAvr) {
				//Take only lower short durations
				continue
			}
			currParam := make(map[string]string)
			//Copy non variable params
			for key, val := range constParam {
				currParam[key] = val
			}
			//Adding avr window durations
			currParam[ShortDur] = shortAvr.String()
			currParam[LongDur] = longAvr.String()
			//Appending to result
			resultMap = append(resultMap, currParam)
		}
	}
	return resultMap, nil
}

func (p *ParamSplitterImpl) convertToRange(limitsStr []string) ([]decimal.Decimal, error) {
	ln := len(limitsStr)
	if ln > 3 {
		p.logger.Error("Wrong number of limits passed: ", limitsStr)
		return nil, errors.NewUnexpectedError("Wrong limit slice")
	}
	if ln == 1 {
		value, err := decimal.NewFromString(limitsStr[0])
		if err != nil {
			return nil, err
		}
		return []decimal.Decimal{value}, nil
	} else {
		start, err := decimal.NewFromString(limitsStr[0])
		if err != nil {
			return nil, err
		}
		end, err := decimal.NewFromString(limitsStr[ln-1])
		if err != nil {
			return nil, err
		}
		var step decimal.Decimal
		if ln == 3 {
			step, err = decimal.NewFromString(limitsStr[1])
			if err != nil {
				return nil, err
			}
		} else {
			step = decimal.NewFromInt(1)
		}
		limits := make([]decimal.Decimal, 0)
		currVal := start
		for currVal.LessThanOrEqual(end) {
			limits = append(limits, currVal)
			currVal = currVal.Add(step)
		}
		return limits, nil
	}
}

func NewParamSplitter(logger *zap.SugaredLogger) stmodel.ParamSplitter {
	return &ParamSplitterImpl{logger: logger}
}
