package dto

type AlgorithmsResponse struct {
	Algorithms []*AlgorithmResponse `json:"algorithms"`
}

type AlgorithmResponse struct {
	AlgorithmID uint              `json:"algorithmID"`
	Params      map[string]string `json:"params"`
	Limits      []*MoneyValue     `json:"limits"`
}
