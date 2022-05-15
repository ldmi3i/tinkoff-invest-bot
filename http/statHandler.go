package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"invest-robot/dto"
	"invest-robot/robot"
	"log"
	"net/http"
)

type StatHandler interface {
	AlgorithmStat(c *gin.Context)
}

type DefaultStatHandler struct {
	api    robot.StatAPI
	logger *zap.SugaredLogger
}

func NewStatHandler(statApi robot.StatAPI, logger *zap.SugaredLogger) StatHandler {
	return &DefaultStatHandler{statApi, logger}
}

func (st *DefaultStatHandler) AlgorithmStat(c *gin.Context) {
	var req dto.StatAlgoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		st.logger.Errorf("Error while validating StatAlgorithm request:\n%s", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	st.logger.Infof("Accept StatAlgoRequest: %+v", req)
	stat, err := st.api.GetAlgorithmStat(&req)
	if err != nil {
		log.Printf("Error while collecting algorithm statistics:\n%s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, stat)
}
