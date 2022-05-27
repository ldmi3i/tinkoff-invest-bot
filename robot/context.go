package robot

import (
	"context"
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

//Initializes application context and populate it with objects
func initConfiguration() {
	logConf := zap.NewDevelopmentConfig()
	if helper.GetLogFilePath() != "" {
		logConf.OutputPaths = []string{
			helper.GetLogFilePath(),
			"stderr",
		}
	}
	logger, err := logConf.Build()
	if err != nil {
		log.Panicf("Error setup logger: %s", err)
	}
	log.Println("Err: ", err)
	sugared := logger.Sugar()
	tapi := tinapi.NewTinApi(sugared)
	infoSdxSrv := service.NewInfoSandboxService(tapi, sugared)
	infoProdSrv := service.NewInfoProdService(tapi, sugared)
	tradeSdxSrv := service.NewTradeSandboxSrv(tapi, sugared)
	tradeProdSrv := service.NewTradeProdService(tapi, sugared)

	hRep := repository.NewHistoryRepository(helper.GetDB())
	actionRep := repository.NewActionRepository(helper.GetDB())
	aRep := repository.NewAlgoRepository(helper.GetDB())
	statRep := repository.NewStatRepository(helper.GetDB())

	statSrv := service.NewStatService(statRep, sugared)
	aFact := strategy.NewAlgFactory(infoSdxSrv, infoProdSrv, hRep, sugared)
	sdxTrader := trade.NewSandboxTrader(infoSdxSrv, tradeSdxSrv, actionRep, sugared)
	prodTrader := trade.NewProdTrader(infoProdSrv, tradeProdSrv, actionRep, sugared)

	ctx = Context{
		infoSdxSrv:   infoSdxSrv,
		infoProdSrv:  infoProdSrv,
		tradeSdxSrv:  tradeSdxSrv,
		tradeProdSrv: tradeProdSrv,
		statSrv:      statSrv,
		hRep:         hRep,
		aRep:         aRep,
		actionRep:    actionRep,
		statRep:      statRep,
		aFact:        aFact,
		sdxTrader:    sdxTrader,
		prodTrader:   prodTrader,
		logger:       sugared,
		ctx:          context.Background(),
	}
}

//Context keeps objects of all API classes.
//Using of Context is preferred way of retrieving instances of all objects.
type Context struct {
	infoSdxSrv   service.InfoSrv      //Sandbox information service
	infoProdSrv  service.InfoSrv      //Prod information service
	tradeSdxSrv  service.TradeService //Sandbox trade service
	tradeProdSrv service.TradeService //Prod trade service
	statSrv      service.StatService
	hRep         repository.HistoryRepository
	aRep         repository.AlgoRepository
	actionRep    repository.ActionRepository
	statRep      repository.StatRepository
	aFact        strategy.AlgFactory
	sdxTrader    trade.Trader //Sandbox trader
	prodTrader   trade.Trader //Prod trader
	logger       *zap.SugaredLogger
	ctx          context.Context
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

func (ctx *Context) GetStatService() service.StatService {
	return ctx.statSrv
}

func Init() {
	helper.InitEnv()
	helper.InitDB()
	helper.InitGRPC()
	initConfiguration()
}

//StartBgTasks start required background tasks.
func StartBgTasks() {
	ctx.logger.Info("Starting background tasks...")
	ctx.sdxTrader.Go(ctx.ctx)  //Starting sandbox trader
	ctx.prodTrader.Go(ctx.ctx) //Starting prod trader
}

func PostProcess() {
	ctx.logger.Info("Sync logs")
	err := ctx.logger.Sync()
	if err != nil {
		return
	}
}
