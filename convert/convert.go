package convert

import (
	"github.com/shopspring/decimal"
	investapi "invest-robot/tapigen"
)

func QuotationToDec(q *investapi.Quotation) decimal.Decimal {
	whole := decimal.NewFromInt(q.Units)
	fr := decimal.New(int64(q.Nano), -9)
	return whole.Add(fr)
}
