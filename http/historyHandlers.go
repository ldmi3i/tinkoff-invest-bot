package http

import (
	"github.com/gin-gonic/gin"
	"invest-robot/dto"
	"invest-robot/robot"
	"log"
	"net/http"
	"time"
)

type HistoryHandler interface {
	LoadHistory(c *gin.Context)
	AnalyzeHistory(c *gin.Context)
	AnalyzeHistoryInRange(c *gin.Context)
}

type DefaultHistoryHandler struct {
	api robot.HistoryAPI
}

func NewHistoryHandler(h robot.HistoryAPI) HistoryHandler {
	return &DefaultHistoryHandler{h}
}

func (h *DefaultHistoryHandler) LoadHistory(c *gin.Context) {
	req := dto.LoadHistoryRequest{Interval: 5}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Printf("Error while validating request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if req.StartTime >= req.EndTime {
		log.Printf("Start time is after then end time...")
		c.JSON(http.StatusBadRequest, "Start time must be after then end time")
		return
	}
	log.Printf("Load history: %+v", req)
	err = h.api.LoadHistory(req.Figis, req.Interval, time.Unix(req.StartTime, 0), time.Unix(req.EndTime, 0))
	if err != nil {
		log.Printf("Error while loading history:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *DefaultHistoryHandler) AnalyzeHistory(c *gin.Context) {
	var req dto.CreateAlgorithmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error while validating AnalyzeAlgo request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("Analyze history: %+v", req)
	stat, err := h.api.AnalyzeAlgo(&req)
	if err != nil {
		log.Printf("Error while analyzing history:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, stat)
}

func (h *DefaultHistoryHandler) AnalyzeHistoryInRange(c *gin.Context) {
	var req dto.CreateAlgorithmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error while validating AnalyzeAlgo request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("Analyze history: %+v", req)
	stat, err := h.api.AnalyzeAlgoInRange(&req)
	if err != nil {
		log.Printf("Error while analyzing history:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, stat)
}
