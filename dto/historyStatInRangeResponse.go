package dto

//HistStatInRangeResponse history range analysis result
type HistStatInRangeResponse struct {
	BestRes *HistStatResponse `json:"bestRes"` //best stat result
	Params  map[string]string `json:"params"`  //best stat result parameters
}
