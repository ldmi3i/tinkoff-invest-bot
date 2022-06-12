package avr

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/collections"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCalcAvr(t *testing.T) {
	tList := collections.NewTList[decimal.Decimal](time.Minute)
	firstHalfVal := decimal.NewFromInt(1)
	secondHalfVal := decimal.NewFromInt(3)
	thirdQuoterVal := decimal.NewFromInt(3)
	startTime := time.Now()
	tList.Append(firstHalfVal, startTime.Add(time.Second*15))
	tList.Append(secondHalfVal, startTime.Add(time.Second*45))
	tList.Append(thirdQuoterVal, startTime.Add(time.Second*55))
	res, err := calcAvr(&tList)
	assert.Nil(t, err)
	assert.True(t, res.GreaterThan(decimal.NewFromFloat(2.2)) && res.LessThan(decimal.NewFromFloat(2.3)))
}

func TestCalcAvr_averageFromTwoEqualsAverage(t *testing.T) {
	tList := collections.NewTList[decimal.Decimal](time.Minute)
	firstHalfVal := decimal.NewFromInt(1)
	secondHalfVal := decimal.NewFromInt(3)

	startTime := time.Now()
	tList.Append(firstHalfVal, startTime.Add(time.Second*15))
	tList.Append(secondHalfVal, startTime.Add(time.Second*45))

	res, err := calcAvr(&tList)
	assert.Nil(t, err)
	expected := decimal.NewFromInt(200)
	assert.Equal(t, res.Mul(decimal.NewFromInt(100)).IntPart(), expected.IntPart())
}
