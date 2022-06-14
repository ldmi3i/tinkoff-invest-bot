package avr

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestConfigure(t *testing.T) {
	config := `{
			"instruments": [
				{
					"figi": "BBG004S681B4",
					"amount": 6,
					"buyPosPrice": 5436
				}
			]
		}`
	confMap := make(map[string]string)
	confMap[dto.InstrAmountField] = config
	var state algoState

	err := configure(confMap, &state, zap.NewExample().Sugar())
	assert.Nil(t, err)
}
