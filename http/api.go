package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"invest-robot/helper"
	"invest-robot/repository"
	"invest-robot/robot"
	"invest-robot/service"
	"invest-robot/strategy"
	"invest-robot/tinapi"
	"log"
)

func StartHttp() {
	router := gin.Default()

	historyHandlers(router)

	log.Fatal(router.Run(fmt.Sprintf("%s:%s", helper.GetSrvHost(), helper.GetSrvPort())))
}

func historyHandlers(router *gin.Engine) {
	infoSrv := service.NewInfoService(tinapi.NewTinApi())
	hRep := repository.NewHistoryRepository(helper.GetDB())
	aFact := strategy.NewAlgFactory(&infoSrv, &hRep)
	aRep := repository.NewAlgoRepository()
	hh := NewHistoryHandler(robot.NewHistoryAPI(infoSrv, hRep, aFact, aRep))

	router.POST("/history/load", hh.LoadHistory)
}
