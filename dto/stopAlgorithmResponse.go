package dto

type StopAlgorithmResponse struct {
	IsStopped bool   `json:"isStopped"`
	Info      string `json:"info"`
}
