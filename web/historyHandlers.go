package web

import (
	"github.com/gin-gonic/gin"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/bot"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type HistoryHandler interface {
	LoadHistory(c *gin.Context)
	AnalyzeHistory(c *gin.Context)
	AnalyzeHistoryInRange(c *gin.Context)
}

type DefaultHistoryHandler struct {
	api    bot.HistoryAPI
	logger *zap.SugaredLogger
}

func NewHistoryHandler(h bot.HistoryAPI, logger *zap.SugaredLogger) HistoryHandler {
	return &DefaultHistoryHandler{h, logger}
}

func (h *DefaultHistoryHandler) LoadHistory(c *gin.Context) {
	req := dto.LoadHistoryRequest{Interval: 5}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Errorf("Error while validating request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if req.StartTime >= req.EndTime {
		h.logger.Error("Start time is after then end time...")
		c.JSON(http.StatusBadRequest, "Start time must be after then end time")
		return
	}
	h.logger.Infof("Load history: %+v", req)
	err = h.api.LoadHistory(req.Figis, req.Interval, time.Unix(req.StartTime, 0), time.Unix(req.EndTime, 0), c.Request.Context())
	if err != nil {
		h.logger.Errorf("Error while loading history:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *DefaultHistoryHandler) AnalyzeHistory(c *gin.Context) {
	var req dto.CreateAlgorithmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Error while validating AnalyzeAlgo request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	h.logger.Infof("Analyze history: %+v", req)
	stat, err := h.api.AnalyzeAlgo(&req, c.Request.Context())
	if err != nil {
		h.logger.Errorf("Error while analyzing history:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, stat)
}

func (h *DefaultHistoryHandler) AnalyzeHistoryInRange(c *gin.Context) {
	var req dto.CreateAlgorithmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Error while validating AnalyzeAlgo request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	h.logger.Infof("Analyze history: %+v", req)
	stat, err := h.api.AnalyzeAlgoInRange(&req, c.Request.Context())
	if err != nil {
		h.logger.Errorf("Error while analyzing history:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, stat)
}
