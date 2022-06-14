package dtotapi

import "github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"

type LastPricesRequest struct {
	Figis []string
}

func (req *LastPricesRequest) ToTinApi() *investapi.GetLastPricesRequest {
	return &investapi.GetLastPricesRequest{Figi: req.Figis}
}
