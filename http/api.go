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

	log.Fatal(router.Run(fmt.Sprintf(":%s", helper.GetSrvPort())))
}

func historyHandlers(router *gin.Engine, ctx *robot.Context) {
	hh := NewHistoryHandler(robot.NewHistoryAPI(
		ctx.GetSandboxInfoSrv(), ctx.GetHistRep(), ctx.GetAlgFactory(), ctx.GetAlgRepository(), ctx.GetLogger()))

	router.POST("/history/load", hh.LoadHistory)
	router.POST("/history/analyze", hh.AnalyzeHistory)
	router.POST("/history/analyze/range", hh.AnalyzeHistoryInRange)
}

func tradeHandlers(router *gin.Engine, ctx *robot.Context) {
	sandboxApi := robot.NewSandboxTradeAPI(ctx.GetSandboxInfoSrv(), ctx.GetAlgFactory(), ctx.GetAlgRepository(),
		ctx.GetSandboxTrader(), ctx.GetLogger())
	prodApi := robot.NewTradeProdAPI(ctx.GetProdInfoSrv(), ctx.GetAlgFactory(), ctx.GetAlgRepository(),
		ctx.GetProdTrader(), ctx.GetLogger())
	th := NewTradeHandler(sandboxApi, prodApi)

	router.POST("/trade/sandbox", th.TradeSandbox)
	router.POST("/trade/prod", th.TradeProd)
}
