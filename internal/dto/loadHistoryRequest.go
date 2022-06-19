package dto

import "github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"

type LoadHistoryRequest struct {
	Figis     []string                 `json:"figis"`
	StartTime int64                    `json:"start_time"` //Start time unix time ms
	EndTime   int64                    `json:"end_time"`   //End time unix time ms
	Interval  investapi.CandleInterval `json:"interval,default=5"`
}
