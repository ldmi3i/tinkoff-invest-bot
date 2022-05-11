package robot

import (
	"invest-robot/domain"
	"invest-robot/dto"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy"
	investapi "invest-robot/tapigen"
	"invest-robot/trade"
	"log"
	"time"
)

type HistoryAPI interface {
	LoadHistory(figis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time) error
	AnalyzeHistory(req *dto.CreateAlgorithmRequest) (*dto.HistStatResponse, error)
}

type DefaultHistoryAPI struct {
	infoSrv service.InfoSrv
	histRep repository.HistoryRepository
	aFact   strategy.AlgFactory
	aRep    repository.AlgoRepository
}

func (h DefaultHistoryAPI) LoadHistory(figis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time) error {
	history, err := h.infoSrv.GetHistorySorted(figis, ivl, startTime, endTime)
	if err != nil {
		return err
	}
	if err = h.histRep.SaveAll(history); err != nil {
		return err
	}
	return nil
}

func (h DefaultHistoryAPI) AnalyzeHistory(req *dto.CreateAlgorithmRequest) (*dto.HistStatResponse, error) {
	algDm := domain.AlgorithmFromDto(req)
	alg, err := h.aFact.NewHist(algDm)
	if err != nil {
		return nil, err
	}
	sub, err := alg.Subscribe()
	if err != nil {
		return nil, err
	}
	shares, err := h.infoSrv.GetAllShares()
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
	trDr := trade.NewMockTrader(h.histRep, lots, figiCurrency)
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

func NewHistoryAPI(infoSrv service.InfoSrv, histRep repository.HistoryRepository, aFact strategy.AlgFactory,
	aRep repository.AlgoRepository) HistoryAPI {
	return DefaultHistoryAPI{infoSrv, histRep, aFact, aRep}
}
