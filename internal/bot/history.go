package bot

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/dto/dtotapi"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/repository"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/service"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/strategy"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/strategy/stmodel"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/trade"
	"go.uber.org/zap"
	"log"
	"sync"
	"time"
)

//HistoryAPI is an interface for interacting with history data
type HistoryAPI interface {
	//LoadHistory loads history from API and replace existing data in database with newly retrieved
	LoadHistory(figis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time, ctx context.Context) error
	//AnalyzeAlgo Analyze algorithm with fixed parameters
	AnalyzeAlgo(req *dto.CreateAlgorithmRequest, ctx context.Context) (*dto.HistStatResponse, error)
	//AnalyzeAlgoInRange Analyze algorithm with parameter variation
	AnalyzeAlgoInRange(req *dto.CreateAlgorithmRequest, ctx context.Context) (*dto.HistStatInRangeResponse, error)
}

type DefaultHistoryAPI struct {
	infoSrv service.InfoSrv
	histRep repository.HistoryRepository
	aFact   strategy.AlgFactory
	aRep    repository.AlgoRepository
	logger  *zap.SugaredLogger
}

func (h *DefaultHistoryAPI) LoadHistory(figis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time, ctx context.Context) error {
	h.logger.Infof("Load history for figis: %s in interval: %d, From: %s to %s", figis, ivl, startTime, endTime)
	history, err := h.infoSrv.GetHistorySorted(figis, ivl, startTime, endTime, ctx)
	if err != nil {
		return err
	}
	if err = h.histRep.ClearAndSaveAll(history); err != nil {
		return err
	}
	h.logger.Infof("Load history completed. Loaded %d entries", len(history))
	return nil
}

func (h *DefaultHistoryAPI) AnalyzeAlgo(req *dto.CreateAlgorithmRequest, ctx context.Context) (*dto.HistStatResponse, error) {
	h.logger.Info("Analyze algorithm request: ", req)
	algDm := entity.AlgorithmFromDto(req)
	alg, err := h.aFact.NewHist(algDm)
	if err != nil {
		return nil, err
	}
	shares, err := h.infoSrv.GetAllShares(ctx)
	if err != nil {
		return nil, err
	}
	res, err := h.performAnalysis(req, shares, alg, ctx)
	if err != nil {
		return nil, err
	}
	h.logger.Info("Analyze algorithm result: ", res)
	return res, nil
}

func (h *DefaultHistoryAPI) performAnalysis(req *dto.CreateAlgorithmRequest, shares *dtotapi.SharesResponse,
	alg stmodel.Algorithm, ctx context.Context) (*dto.HistStatResponse, error) {
	sub, err := alg.Subscribe()
	if err != nil {
		return nil, err
	}
	figiSet := make(map[string]bool)
	for _, figi := range req.Figis {
		figiSet[figi] = true
	}
	lots := make(map[string]int64)
	figiCurrency := make(map[string]string)
	for _, instr := range shares.Instruments {
		if !figiSet[instr.Figi] {
			continue
		}
		lots[instr.Figi] = instr.Lot
		figiCurrency[instr.Figi] = instr.Currency
	}
	trDr := trade.NewMockTrader(h.histRep, lots, figiCurrency, h.logger)
	if err = trDr.AddSubscription(sub); err != nil {
		return nil, err
	}
	trDr.Go(ctx)
	if err = alg.Go(ctx); err != nil {
		log.Printf("Error while starting algorithm")
		return nil, err
	}
	res := <-trDr.GetStatCh()
	return &res, nil
}

func (h *DefaultHistoryAPI) AnalyzeAlgoInRange(req *dto.CreateAlgorithmRequest, ctx context.Context) (*dto.HistStatInRangeResponse, error) {
	h.logger.Info("Analyze algorithm in range from request: ", req)
	algDm := entity.AlgorithmFromDto(req)
	algRange, err := h.aFact.NewRange(algDm)
	if err != nil {
		return nil, err
	}
	shares, err := h.infoSrv.GetAllShares(ctx)
	if err != nil {
		return nil, err
	}

	resCh := h.rangeAnalyzeBg(algRange, req, shares, ctx)
	analyzeRes := make([]*dto.HistStatIdDto, 0, len(algRange))

	for algRes := range resCh {
		h.logger.Debug("Result received: ", algRes)
		analyzeRes = append(analyzeRes, algRes)
	}

	if len(analyzeRes) == 0 {
		return nil, nil
	}
	var top = analyzeRes[0]
	for _, stat := range analyzeRes {
		if top.HistStat.CurBalance["rub"].LessThan(stat.HistStat.CurBalance["rub"]) {
			top = stat
		}
	}
	res := &dto.HistStatInRangeResponse{
		BestRes: top.HistStat,
		Params:  top.Param,
	}
	h.logger.Info("Analyze algorithm in range completed; Result: ", res)
	return res, nil
}

func (h *DefaultHistoryAPI) rangeAnalyzeBg(algRange []stmodel.Algorithm, req *dto.CreateAlgorithmRequest,
	shares *dtotapi.SharesResponse, ctx context.Context) chan *dto.HistStatIdDto {
	var wg sync.WaitGroup
	resCh := make(chan *dto.HistStatIdDto)
	concurrency := 18
	semaphore := make(chan bool, concurrency)
	//Performs background algorithm processing
	go func() {
	OUT:
		for _, alg := range algRange {
			select {
			case <-ctx.Done():
				h.logger.Infof("History range context closed with err: %s, stopping...", ctx.Err())
				break OUT
			case semaphore <- true:
				wg.Add(1)
				go func(alg stmodel.Algorithm) {
					defer func() {
						<-semaphore
						wg.Done()
					}()
					histResult, err := h.performAnalysis(req, shares, alg, ctx)
					if err != nil {
						h.logger.Error("Error while performing algorithm analysis: ", err)
						return
					}
					resCh <- &dto.HistStatIdDto{
						Id:       alg.GetAlgorithm().ID,
						HistStat: histResult,
						Param:    alg.GetParam(),
					}
				}(alg)
			}
		}
		wg.Wait()
		close(resCh)
	}()

	return resCh
}

func NewHistoryAPI(infoSrv service.InfoSrv, histRep repository.HistoryRepository, aFact strategy.AlgFactory,
	aRep repository.AlgoRepository, logger *zap.SugaredLogger) HistoryAPI {
	return &DefaultHistoryAPI{infoSrv, histRep, aFact, aRep, logger}
}
