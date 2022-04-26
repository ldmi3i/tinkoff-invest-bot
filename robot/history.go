package robot

import (
	"invest-robot/helper"
	"invest-robot/service"
	investapi "invest-robot/tapigen"
	"time"
)

type HistoryAPI interface {
	LoadHistory(figis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time) error
}

type DefaultHistoryAPI struct {
	infoSrv service.InfoSrv
}

func (h DefaultHistoryAPI) LoadHistory(figis []string, ivl investapi.CandleInterval, startTime time.Time, endTime time.Time) error {
	history, err := h.infoSrv.GetHistory(figis, ivl, startTime, endTime)
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
