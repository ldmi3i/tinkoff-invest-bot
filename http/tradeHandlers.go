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
}

type DefaultTradeHandler struct {
	api robot.TradeAPI
}

func NewTradeHandler(api robot.TradeAPI) TradeHandler {
	return DefaultTradeHandler{api}
}

func (h DefaultTradeHandler) TradeSandbox(c *gin.Context) {
	var req dto.CreateAlgorithmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error while validating AnalyzeHistory request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("Analyze history: %+v", req)
	stat, err := h.api.TradeSandbox(req)
	if err != nil {
		log.Printf("Error while analyzing history:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, stat)
}
