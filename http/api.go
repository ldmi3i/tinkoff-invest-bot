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

type apiCtx struct {
	infoSrv service.InfoSrv
	hRep    repository.HistoryRepository
	aFact   strategy.AlgFactory
	aRep    repository.AlgoRepository
}

func StartHttp() {
	router := gin.Default()

	infoSrv := service.NewInfoService(tinapi.NewTinApi())
	hRep := repository.NewHistoryRepository(helper.GetDB())
	aFact := strategy.NewAlgFactory(infoSrv, hRep)
	aRep := repository.NewAlgoRepository()

	ctx := apiCtx{
		infoSrv: infoSrv,
		hRep:    hRep,
		aFact:   aFact,
		aRep:    aRep,
	}

	historyHandlers(router, &ctx)

	log.Fatal(router.Run(fmt.Sprintf("%s:%s", helper.GetSrvHost(), helper.GetSrvPort())))
}

func historyHandlers(router *gin.Engine, ctx *apiCtx) {
	hh := NewHistoryHandler(robot.NewHistoryAPI(ctx.infoSrv, ctx.hRep, ctx.aFact, ctx.aRep))

	router.POST("/history/load", hh.LoadHistory)
	router.POST("/history/analyze", hh.AnalyzeHistory)

}

func tradeHandlers(router *gin.Engine, ctx *apiCtx) {
	th := NewTradeHandler(robot.NewSandboxTradeAPI(ctx.infoSrv, ctx.hRep, ctx.aFact, ctx.aRep))

	router.POST("/trade/add", th.TradeSandbox)
}
