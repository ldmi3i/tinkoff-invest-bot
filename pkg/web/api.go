package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"invest-robot/internal/env"
	"invest-robot/pkg/robot"
	"log"
)

func StartHttp() {
	router := gin.Default()
	ctx := robot.GetContext()
	historyHandlers(router, ctx)
	tradeHandlers(router, ctx)
	statHandlers(router, ctx)

	log.Fatal(router.Run(fmt.Sprintf(":%s", env.GetSrvPort())))
}

func historyHandlers(router *gin.Engine, ctx *robot.Context) {
	hh := NewHistoryHandler(robot.NewHistoryAPI(
		ctx.GetSandboxInfoSrv(), ctx.GetHistRep(), ctx.GetAlgFactory(), ctx.GetAlgRepository(), ctx.GetLogger()),
		ctx.GetLogger(),
	)

	router.POST("/history/load", hh.LoadHistory)
	router.POST("/history/analyze", hh.AnalyzeHistory)
	router.POST("/history/analyze/range", hh.AnalyzeHistoryInRange)
}

func tradeHandlers(router *gin.Engine, ctx *robot.Context) {
	sandboxApi := robot.NewSandboxTradeAPI(ctx.GetSandboxInfoSrv(), ctx.GetAlgFactory(), ctx.GetAlgRepository(),
		ctx.GetSandboxTrader(), ctx.GetLogger())
	prodApi := robot.NewTradeProdAPI(ctx.GetProdInfoSrv(), ctx.GetAlgFactory(), ctx.GetAlgRepository(),
		ctx.GetProdTrader(), ctx.GetLogger())
	th := NewTradeHandler(sandboxApi, prodApi, ctx.GetLogger())

	router.POST("/trade/sandbox", th.TradeSandbox)
	router.POST("/trade/prod", th.TradeProd)

	router.GET("/trade/algorithms/active/prod", th.GetProdAlgorithms)
	router.GET("/trade/algorithms/active/sandbox", th.GetSdbxAlgorithms)
	router.POST("/trade/algorithms/stop", th.StopAlgorithm)
}

func statHandlers(router *gin.Engine, ctx *robot.Context) {
	statApi := robot.NewStatAPI(ctx.GetStatService(), ctx.GetLogger())
	st := NewStatHandler(statApi, ctx.GetLogger())

	router.GET("/stat/algorithm", st.AlgorithmStat)
}
