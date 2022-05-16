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
	GetSdbxAlgorithms(c *gin.Context)
	GetProdAlgorithms(c *gin.Context)
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

func (h *DefaultTradeHandler) GetSdbxAlgorithms(c *gin.Context) {
	h.logger.Info("GetProdAlgorithms sandbox requested")
	algos, err := h.sandboxApi.GetActiveAlgorithms()
	if err != nil {
		h.logger.Errorf("Error retrieving active algorithms:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	h.logger.Infof("GetProdAlgorithms sandbox returns %d results", len(algos.Algorithms))
	c.JSON(http.StatusOK, algos)
}

func (h *DefaultTradeHandler) GetProdAlgorithms(c *gin.Context) {
	h.logger.Info("GetProdAlgorithms prod requested")
	algos, err := h.prodApi.GetActiveAlgorithms()
	if err != nil {
		h.logger.Errorf("Error retrieving active algorithms:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	h.logger.Infof("GetProdAlgorithms prod returns %d results", len(algos.Algorithms))
	c.JSON(http.StatusOK, algos)
}
