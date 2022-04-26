package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"invest-robot/helper"
	"invest-robot/robot"
	"invest-robot/service"
	"invest-robot/tinapi"
	"log"
)

func StartHttp() {
	router := gin.Default()

	historyHandlers(router)

	log.Fatal(router.Run(fmt.Sprintf("%s:%s", helper.GetSrvHost(), helper.GetSrvPort())))
}

func historyHandlers(router *gin.Engine) {
	hh := NewHistoryHandler(
		robot.NewHistoryAPI(
			service.NewInfoService(tinapi.NewTinApi()),
		),
	)

	router.POST("/history/load", hh.LoadHistory)
}
