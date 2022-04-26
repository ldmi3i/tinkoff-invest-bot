package robot

import (
	"invest-robot/helper"
	"invest-robot/service"
	investapi "invest-robot/tapigen"
	"time"
)

type HistoryAPI interface {
	LoadHistory(figis []string, startTime time.Time, endTime time.Time) error
}

type DefaultHistoryAPI struct {
	infoSrv service.InfoSrv
}

func (h DefaultHistoryAPI) LoadHistory(figis []string, startTime time.Time, endTime time.Time) error {
	history, err := h.infoSrv.GetHistory(figis, investapi.CandleInterval_CANDLE_INTERVAL_DAY, startTime, endTime)
	if err != nil {
		return err
	}
	db := helper.GetDB()

	db.Exec("DELETE FROM history")
	db.Create(&history)
	return nil
}

func NewHistoryAPI(infoSrv service.InfoSrv) HistoryAPI {
	return DefaultHistoryAPI{infoSrv}
}
