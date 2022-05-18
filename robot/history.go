package robot

import (
	"go.uber.org/zap"
	"invest-robot/domain"
	"invest-robot/dto"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy"
	"invest-robot/strategy/stmodel"
	investapi "invest-robot/tapigen"
	"invest-robot/trade"
	"log"
	"sync"
	"time"
)

type HistoryAPI interface {
	LoadHistory(figis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time) error
	AnalyzeAlgo(req *dto.CreateAlgorithmRequest) (*dto.HistStatResponse, error)
	AnalyzeAlgoInRange(req *dto.CreateAlgorithmRequest) (*dto.HistStatInRangeResponse, error)
}

type DefaultHistoryAPI struct {
	infoSrv service.InfoSrv
	histRep repository.HistoryRepository
	aFact   strategy.AlgFactory
	aRep    repository.AlgoRepository
	logger  *zap.SugaredLogger
}

func (h *DefaultHistoryAPI) LoadHistory(figis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time) error {
	h.logger.Infof("Load history for figis: %s in interval: %d, From: %s to %s", figis, ivl, startTime, endTime)
	history, err := h.infoSrv.GetHistorySorted(figis, ivl, startTime, endTime)
	if err != nil {
		return err
	}
	if err = h.histRep.SaveAll(history); err != nil {
		return err
	}
	h.logger.Infof("Load history completed. Loaded %d entries", len(history))
	return nil
}

func (h *DefaultHistoryAPI) AnalyzeAlgo(req *dto.CreateAlgorithmRequest) (*dto.HistStatResponse, error) {
	h.logger.Info("Analyze algorithm request: ", req)
	algDm := domain.AlgorithmFromDto(req)
	alg, err := h.aFact.NewHist(algDm)
	if err != nil {
		return nil, err
	}
	shares, err := h.infoSrv.GetAllShares()
	if err != nil {
		return nil, err
	}
	res, err := h.performAnalysis(req, shares, alg)
	if err != nil {
		return nil, err
	}
	h.logger.Info("Analyze algorithm result: ", res)
	return res, nil
}

func (h *DefaultHistoryAPI) performAnalysis(req *dto.CreateAlgorithmRequest, shares *investapi.SharesResponse,
	alg stmodel.Algorithm) (*dto.HistStatResponse, error) {
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
		lots[instr.Figi] = int64(instr.Lot)
		figiCurrency[instr.Figi] = instr.Currency
	}
	trDr := trade.NewMockTrader(h.histRep, lots, figiCurrency, h.logger)
	if err = trDr.AddSubscription(sub); err != nil {
		return nil, err
	}
	trDr.Go()
	if err = alg.Go(); err != nil {
		log.Printf("Error while starting algorithm, check routine leaking")
		return nil, err
	}
	res := <-trDr.GetStatCh()
	return &res, nil
}

func (h *DefaultHistoryAPI) AnalyzeAlgoInRange(req *dto.CreateAlgorithmRequest) (*dto.HistStatInRangeResponse, error) {
	h.logger.Info("Analyze algorithm in range from request: ", req)
	algDm := domain.AlgorithmFromDto(req)
	algRange, err := h.aFact.NewRange(algDm)
	if err != nil {
		return nil, err
	}
	shares, err := h.infoSrv.GetAllShares()
	if err != nil {
		return nil, err
	}

	doneCh, resCh := h.rangeAnalyzeBg(algRange, req, shares)
	analyzeRes := make([]*dto.HistStatIdDto, 0, len(algRange))
OUT:
	for {
		select {
		case algRes := <-resCh:
			h.logger.Debug("Result received: ", algRes)
			analyzeRes = append(analyzeRes, algRes)
		case <-doneCh:
			close(doneCh)
			close(resCh)
			break OUT
		}
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
	shares *investapi.SharesResponse) (chan bool, chan *dto.HistStatIdDto) {
	var wg sync.WaitGroup
	resCh := make(chan *dto.HistStatIdDto)
	concurrency := 18
	semaphore := make(chan bool, concurrency)
	//Performs background algorithm processing
	for _, alg := range algRange {
		wg.Add(1)
		go func(alg stmodel.Algorithm) {
			semaphore <- true
			defer func() {
				<-semaphore
				wg.Done()
			}()
			histResult, err := h.performAnalysis(req, shares, alg)
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
	//Make channel indicating result (in case of errors result len may not be equals to algRange len)
	doneCh := make(chan bool)
	//Launch in bg populating done channel when all groups completed
	go func() {
		wg.Wait()
		doneCh <- true
	}()
	return doneCh, resCh
}

func NewHistoryAPI(infoSrv service.InfoSrv, histRep repository.HistoryRepository, aFact strategy.AlgFactory,
	aRep repository.AlgoRepository, logger *zap.SugaredLogger) HistoryAPI {
	return &DefaultHistoryAPI{infoSrv, histRep, aFact, aRep, logger}
}
