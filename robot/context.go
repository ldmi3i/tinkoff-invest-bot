package robot

import (
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
	tapi := tinapi.NewTinApi()
	infoSrv := service.NewSandboxInfoService(tapi)
	sdxTradeSrv := service.NewSandboxTradeSrv(tapi)
	hRep := repository.NewHistoryRepository(helper.GetDB())
	actionRep := repository.NewActionRepository(helper.GetDB())
	aFact := strategy.NewAlgFactory(infoSrv, hRep)
	aRep := repository.NewAlgoRepository()
	sdxTrader := trade.NewSandboxTrader(infoSrv, sdxTradeSrv, actionRep)
	prodTrader := trade.NewProdApiTrader()

	ctx = Context{
		infoSrv:     infoSrv,
		sdxTradeSrv: sdxTradeSrv,
		hRep:        hRep,
		aRep:        aRep,
		actionRep:   actionRep,
		aFact:       aFact,
		sdxTrader:   sdxTrader,
		prodTrader:  prodTrader,
	}
}

type Context struct {
	infoSrv     service.InfoSrv
	sdxTradeSrv service.TradeService
	hRep        repository.HistoryRepository
	aRep        repository.AlgoRepository
	actionRep   repository.ActionRepository
	aFact       strategy.AlgFactory
	sdxTrader   trade.Trader
	prodTrader  trade.Trader
}

func (ctx *Context) GetSandboxInfoSrv() service.InfoSrv {
	return ctx.infoSrv
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

func StartBgTasks() {
	log.Println("Starting background tasks...")
	ctx.sdxTrader.Go()
	//TODO add prod trader bg start
}
