package convert

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertingQuotation(t *testing.T) {
	init, err := decimal.NewFromString("1234567890.987654320") //must be 9 after . for same exponent
	assert.Nil(t, err)
	qt := DecToQuotation(init)
	assert.NotNil(t, qt)
	dec := QuotationToDec(qt)
	assert.Equal(t, init, dec)
}
