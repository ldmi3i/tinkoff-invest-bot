package dto

type HistStatInRangeResponse struct {
	BestRes *HistStatResponse `json:"bestRes"`
	Params  map[string]string `json:"params"`
}
