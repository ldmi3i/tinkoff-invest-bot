package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ldmi3i/tinkoff-invest-bot/bot"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/env"
	"log"
)

func StartHttp() {
	router := gin.Default()
	dc := bot.GetDepContainer()
	historyHandlers(router, dc)
	tradeHandlers(router, dc)
	statHandlers(router, dc)

	log.Fatal(router.Run(fmt.Sprintf(":%s", env.GetSrvPort())))
}

func historyHandlers(router *gin.Engine, ctx bot.DependencyContainer) {
	hh := NewHistoryHandler(ctx.GetHistoryAPI(), ctx.GetLogger())

	router.POST("/history/load", hh.LoadHistory)
	router.POST("/history/analyze", hh.AnalyzeHistory)
	router.POST("/history/analyze/range", hh.AnalyzeHistoryInRange)
}

func tradeHandlers(router *gin.Engine, dc bot.DependencyContainer) {
	sandboxApi := dc.GetSdxTradeAPI()
	prodApi := dc.GetProdTradeAPI()
	th := NewTradeHandler(sandboxApi, prodApi, dc.GetLogger())

	router.POST("/trade/sandbox", th.TradeSandbox)
	router.POST("/trade/prod", th.TradeProd)

	router.GET("/trade/algorithms/active/prod", th.GetProdAlgorithms)
	router.GET("/trade/algorithms/active/sandbox", th.GetSdbxAlgorithms)
	router.POST("/trade/algorithms/stop", th.StopAlgorithm)
}

func statHandlers(router *gin.Engine, dc bot.DependencyContainer) {
	statApi := dc.GetStatAPI()
	st := NewStatHandler(statApi, dc.GetLogger())

	router.GET("/stat/algorithm", st.AlgorithmStat)
}
