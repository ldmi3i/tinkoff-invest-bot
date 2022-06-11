package dtotapi

import investapi "invest-robot/internal/tapigen"

type LastPricesRequest struct {
	Figis []string
}

func (req *LastPricesRequest) ToTinApi() *investapi.GetLastPricesRequest {
	return &investapi.GetLastPricesRequest{Figi: req.Figis}
}
