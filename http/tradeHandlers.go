package http

import (
	"github.com/gin-gonic/gin"
	"invest-robot/dto"
	"invest-robot/robot"
	"log"
	"net/http"
)

type TradeHandler interface {
	TradeSandbox(c *gin.Context)
	TradeProd(c *gin.Context)
}

type DefaultTradeHandler struct {
	sandboxApi robot.TradeAPI
	prodApi    robot.TradeAPI
}

func NewTradeHandler(sandboxApi robot.TradeAPI, prodApi robot.TradeAPI) TradeHandler {
	return &DefaultTradeHandler{sandboxApi, prodApi}
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
		log.Printf("Error while validating CreateAlgorithm request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("Start sandbox trading: %+v", req)
	stat, err := api.Trade(&req)
	if err != nil {
		log.Printf("Error while analyzing history:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, stat)
}
