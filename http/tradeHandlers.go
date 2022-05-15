package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"invest-robot/dto"
	"invest-robot/robot"
	"net/http"
)

type TradeHandler interface {
	TradeSandbox(c *gin.Context)
	TradeProd(c *gin.Context)
}

type DefaultTradeHandler struct {
	sandboxApi robot.TradeAPI
	prodApi    robot.TradeAPI
	logger     *zap.SugaredLogger
}

func NewTradeHandler(sandboxApi robot.TradeAPI, prodApi robot.TradeAPI, logger *zap.SugaredLogger) TradeHandler {
	return &DefaultTradeHandler{sandboxApi, prodApi, logger}
}

func (h *DefaultTradeHandler) TradeSandbox(c *gin.Context) {
	h.trade(c, h.sandboxApi)
}

func (h *DefaultTradeHandler) TradeProd(c *gin.Context) {
	h.trade(c, h.prodApi)
}

func (h *DefaultTradeHandler) trade(c *gin.Context, api robot.TradeAPI) {
	var req dto.CreateAlgorithmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Error while validating CreateAlgorithm request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	h.logger.Infof("Start sandbox trading: %+v", req)
	stat, err := api.Trade(&req)
	if err != nil {
		h.logger.Errorf("Error while analyzing history:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, stat)
}
