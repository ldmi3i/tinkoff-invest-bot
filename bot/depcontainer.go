package bot

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/bot"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/connections/db"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/connections/grpc"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/env"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/repository"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/service"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/strategy"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tinapi"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/trade"
	"github.com/tevino/abool/v2"
	"go.uber.org/zap"
	"log"
)

type DependencyContainer interface {
	//GetLogger returns root logger used by API
	GetLogger() *zap.SugaredLogger
	//GetStatAPI returns statistics API instance
	GetStatAPI() bot.StatAPI
	//GetHistoryAPI returns history API instance
	GetHistoryAPI() bot.HistoryAPI
	//GetSdxTradeAPI returns sandbox trade API instance
	GetSdxTradeAPI() bot.TradeAPI
	//GetProdTradeAPI returns production trade API instance
	GetProdTradeAPI() bot.TradeAPI
}

var dc depContainerImpl
var isInitialized = abool.NewBool(false)

func GetDepContainer() DependencyContainer {
	if isInitialized.IsSet() {
		return &dc
	} else {
		panic("Not initialized container called. " +
			"You need at first call initConfiguration or initConfigurationWithLogger method to populate dependencies.")
	}
}

func initConfiguration() {
	logConf := zap.NewDevelopmentConfig()
	if env.GetLogFilePath() != "" {
		logConf.OutputPaths = []string{
			env.GetLogFilePath(),
			"stderr",
		}
	}
	logger, err := logConf.Build()
	if err != nil {
		log.Panicf("Error setup logger: %s", err)
	}
	log.Println("Err: ", err)
	initConfigurationWithLogger(logger.Sugar())
}

//Initializes application context and populate it with objects
func initConfigurationWithLogger(sugared *zap.SugaredLogger) {
	tapi := tinapi.NewTinApi(sugared)
	infoSdxSrv := service.NewInfoSandboxService(tapi, sugared)
	infoProdSrv := service.NewInfoProdService(tapi, sugared)
	tradeSdxSrv := service.NewTradeSandboxSrv(tapi, sugared)
	tradeProdSrv := service.NewTradeProdService(tapi, sugared)

	hRep := repository.NewHistoryRepository(db.GetDB())
	actionRep := repository.NewActionRepository(db.GetDB())
	aRep := repository.NewAlgoRepository(db.GetDB())
	statRep := repository.NewStatRepository(db.GetDB())

	statSrv := service.NewStatService(statRep, sugared)
	aFact := strategy.NewAlgFactory(infoSdxSrv, infoProdSrv, hRep, sugared)
	sdxTrader := trade.NewSandboxTrader(infoSdxSrv, tradeSdxSrv, actionRep, sugared)
	prodTrader := trade.NewProdTrader(infoProdSrv, tradeProdSrv, actionRep, sugared)

	historyAPI := bot.NewHistoryAPI(infoSdxSrv, hRep, aFact, aRep, sugared)
	sdxTradeAPI := bot.NewSandboxTradeAPI(infoSdxSrv, aFact, aRep, sdxTrader, sugared)
	prodTradeAPI := bot.NewTradeProdAPI(infoProdSrv, aFact, aRep, prodTrader, sugared)
	statAPI := bot.NewStatAPI(statSrv, sugared)

	dc = depContainerImpl{
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
		historyAPI:   historyAPI,
		sdxTradeAPI:  sdxTradeAPI,
		prodTradeAPI: prodTradeAPI,
		statAPI:      statAPI,
	}
}

//depContainerImpl keeps objects of all API classes.
//Using of depContainerImpl is preferred way of retrieving instances of all objects.
type depContainerImpl struct {
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

	statAPI      bot.StatAPI
	historyAPI   bot.HistoryAPI
	sdxTradeAPI  bot.TradeAPI
	prodTradeAPI bot.TradeAPI
}

func (dc *depContainerImpl) GetLogger() *zap.SugaredLogger {
	return dc.logger
}

func (dc *depContainerImpl) GetStatAPI() bot.StatAPI {
	return dc.statAPI
}

func (dc *depContainerImpl) GetHistoryAPI() bot.HistoryAPI {
	return dc.historyAPI
}

func (dc *depContainerImpl) GetSdxTradeAPI() bot.TradeAPI {
	return dc.sdxTradeAPI
}

func (dc *depContainerImpl) GetProdTradeAPI() bot.TradeAPI {
	return dc.prodTradeAPI
}

func Init() {
	if isInitialized.SetToIf(false, true) {
		//If data not initialized
		env.InitEnv()
		db.InitDB()
		grpc.InitGRPC()
		initConfiguration()
	} else {
		//If already initialized
		dc.logger.Warn("Init called more than once, do nothing...")
	}
}

func InitWithLogger(logger *zap.SugaredLogger) {
	if isInitialized.SetToIf(false, true) {
		env.InitEnv()
		db.InitDB()
		grpc.InitGRPC()
		initConfigurationWithLogger(logger)
	} else {
		//If already initialized
		logger.Warn("Init called more than once, do nothing...")
	}
}

//StartBgTasks start required background tasks.
func StartBgTasks() {
	if isInitialized.IsNotSet() {
		panic("Not initialized container called. " +
			"You need at first call initConfiguration or initConfigurationWithLogger method to populate dependencies.")
	}
	dc.logger.Info("Starting background tasks...")
	dc.sdxTrader.Go(dc.ctx)  //Starting sandbox trader
	dc.prodTrader.Go(dc.ctx) //Starting prod trader
}

func PostProcess() {
	err := grpc.Close()
	if err != nil {
		log.Print("error while closing grpc connection:", err)
	}
	if isInitialized.IsSet() {
		dc.logger.Info("Sync logs")
		err = dc.logger.Sync()
		if err != nil {
			log.Print("error while sync logger ", err)
		}
	}
}
