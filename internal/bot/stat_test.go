package bot

import (
	"github.com/golang/mock/gomock"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/mocks/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GetAlgorithmStat_should_delegate_to_service(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := service.NewMockStatService(ctrl)
	req := &dto.StatAlgoRequest{AlgorithmID: 43}
	resp := &dto.StatAlgoResponse{AlgorithmID: 43}
	mockService.EXPECT().GetAlgorithmStat(req).Return(resp, nil)

	statAPI := DefaultStatAPI{
		statSrv: mockService,
		logger:  nil,
	}
	stat, err := statAPI.GetAlgorithmStat(req)
	assert.Nil(t, err)
	assert.Equal(t, resp, stat)
}
