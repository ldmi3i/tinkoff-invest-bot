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
}

type DefaultHistoryHandler struct {
	api robot.HistoryAPI
}

func NewHistoryHandler(h robot.HistoryAPI) HistoryHandler {
	return DefaultHistoryHandler{h}
}

func (h DefaultHistoryHandler) LoadHistory(c *gin.Context) {
	req := dto.LoadHistoryRequest{Interval: 5}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Printf("Error while validating request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
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
