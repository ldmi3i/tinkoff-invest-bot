package dto

type LoadHistoryRequest struct {
	Figis     []string `json:"figis"`
	StartTime int64    `json:"start_time"`
	EndTime   int64    `json:"end_time"`
}
