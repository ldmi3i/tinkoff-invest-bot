package web

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/bot"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"go.uber.org/zap"
	"net/http"
)

type TradeHandler interface {
	TradeSandbox(c *gin.Context)
	TradeProd(c *gin.Context)
	GetSdbxAlgorithms(c *gin.Context)
	GetProdAlgorithms(c *gin.Context)
	StopAlgorithm(c *gin.Context)
}

type DefaultTradeHandler struct {
	sandboxApi bot.TradeAPI
	prodApi    bot.TradeAPI
	logger     *zap.SugaredLogger
}

func NewTradeHandler(sandboxApi bot.TradeAPI, prodApi bot.TradeAPI, logger *zap.SugaredLogger) TradeHandler {
	return &DefaultTradeHandler{sandboxApi, prodApi, logger}
}

func (h *DefaultTradeHandler) StopAlgorithm(c *gin.Context) {
	var req dto.StopAlgorithmRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Errorf("Error while validating Stop Algorithm request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	h.logger.Infof("Start sandbox trading: %+v", req)
	stat, err := h.prodApi.StopAlgorithm(&req)
	if err != nil {
		h.logger.Errorf("Error while stopping algorithm:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, stat)
}

func (h *DefaultTradeHandler) TradeSandbox(c *gin.Context) {
	h.trade(c, h.sandboxApi)
}

func (h *DefaultTradeHandler) TradeProd(c *gin.Context) {
	h.trade(c, h.prodApi)
}

func (h *DefaultTradeHandler) trade(c *gin.Context, api bot.TradeAPI) {
	var req dto.CreateAlgorithmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Error while validating CreateAlgorithm request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	context.Background()
	h.logger.Infof("Start sandbox trading: %+v", req)
	stat, err := api.Trade(&req, context.Background())
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
