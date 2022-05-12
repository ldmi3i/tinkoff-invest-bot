package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"invest-robot/helper"
	"invest-robot/robot"
	"log"
)

func StartHttp() {
	router := gin.Default()
	ctx := robot.GetContext()
	historyHandlers(router, ctx)
	tradeHandlers(router, ctx)

	log.Fatal(router.Run(fmt.Sprintf("%s:%s", helper.GetSrvHost(), helper.GetSrvPort())))
}

func historyHandlers(router *gin.Engine, ctx *robot.Context) {
	hh := NewHistoryHandler(robot.NewHistoryAPI(ctx.GetSandboxInfoSrv(), ctx.GetHistRep(), ctx.GetAlgFactory(), ctx.GetAlgRepository()))

	router.POST("/history/load", hh.LoadHistory)
	router.POST("/history/analyze", hh.AnalyzeHistory)
}

func tradeHandlers(router *gin.Engine, ctx *robot.Context) {
	th := NewTradeHandler(robot.NewSandboxTradeAPI(ctx.GetSandboxInfoSrv(), ctx.GetAlgFactory(), ctx.GetAlgRepository(),
		ctx.GetSandboxTrader(), ctx.GetProdTrader(), ctx.GetLogger()))

	router.POST("/trade/sandbox", th.TradeSandbox)
}
