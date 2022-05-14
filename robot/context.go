package robot

import (
	"go.uber.org/zap"
	"invest-robot/helper"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy"
	"invest-robot/tinapi"
	"invest-robot/trade"
	"log"
)

var ctx Context

func GetContext() *Context {
	return &ctx
}

func init() {
	logger, _ := zap.NewDevelopment()
	sugared := logger.Sugar()
	tapi := tinapi.NewTinApi()
	infoSdxSrv := service.NewInfoSandboxService(tapi, sugared)
	infoProdSrv := service.NewInfoProdService(tapi, sugared)
	tradeSdxSrv := service.NewTradeSandboxSrv(tapi, sugared)
	tradeProdSrv := service.NewTradeProdService(tapi, sugared)
	hRep := repository.NewHistoryRepository(helper.GetDB())
	actionRep := repository.NewActionRepository(helper.GetDB())
	aFact := strategy.NewAlgFactory(infoSdxSrv, hRep, sugared)
	aRep := repository.NewAlgoRepository(helper.GetDB())
	sdxTrader := trade.NewSandboxTrader(infoSdxSrv, tradeSdxSrv, actionRep, sugared)
	prodTrader := trade.NewProdApiTrader(sugared)

	ctx = Context{
		infoSdxSrv:   infoSdxSrv,
		infoProdSrv:  infoProdSrv,
		tradeSdxSrv:  tradeSdxSrv,
		tradeProdSrv: tradeProdSrv,
		hRep:         hRep,
		aRep:         aRep,
		actionRep:    actionRep,
		aFact:        aFact,
		sdxTrader:    sdxTrader,
		prodTrader:   prodTrader,
		logger:       sugared,
	}
}

type Context struct {
	infoSdxSrv   service.InfoSrv
	infoProdSrv  service.InfoSrv
	tradeSdxSrv  service.TradeService
	tradeProdSrv service.TradeService
	hRep         repository.HistoryRepository
	aRep         repository.AlgoRepository
	actionRep    repository.ActionRepository
	aFact        strategy.AlgFactory
	sdxTrader    trade.Trader
	prodTrader   trade.Trader
	logger       *zap.SugaredLogger
}

func (ctx *Context) GetSandboxInfoSrv() service.InfoSrv {
	return ctx.infoSdxSrv
}

func (ctx *Context) GetProdInfoSrv() service.InfoSrv {
	return ctx.infoProdSrv
}

func (ctx *Context) GetHistRep() repository.HistoryRepository {
	return ctx.hRep
}

func (ctx *Context) GetAlgFactory() strategy.AlgFactory {
	return ctx.aFact
}

func (ctx *Context) GetAlgRepository() repository.AlgoRepository {
	return ctx.aRep
}

func (ctx *Context) GetSandboxTrader() trade.Trader {
	return ctx.sdxTrader
}

func (ctx *Context) GetProdTrader() trade.Trader {
	return ctx.prodTrader
}

func (ctx *Context) GetLogger() *zap.SugaredLogger {
	return ctx.logger
}

func StartBgTasks() {
	log.Println("Starting background tasks...")
	ctx.sdxTrader.Go()
	//TODO add prod trader bg start
}
