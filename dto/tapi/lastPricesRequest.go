package tapi

import investapi "invest-robot/tapigen"

type LastPricesRequest struct {
	Figis []string
}

func (req *LastPricesRequest) ToTinApi() *investapi.GetLastPricesRequest {
	return &investapi.GetLastPricesRequest{Figi: req.Figis}
}
