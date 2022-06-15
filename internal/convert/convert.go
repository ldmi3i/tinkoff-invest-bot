package convert

import (
	"github.com/ldmi3i/tinkoff-invest-bot/internal/tapigen"
	"github.com/shopspring/decimal"
)

func QuotationToDec(q *investapi.Quotation) decimal.Decimal {
	if q == nil {
		return decimal.Zero
	}
	whole := decimal.NewFromInt(q.Units)
	fr := decimal.New(int64(q.Nano), -9)
	return whole.Add(fr)
}

func DecToQuotation(d decimal.Decimal) *investapi.Quotation {
	nano := d.Mod(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(1_000_000_000)).IntPart()
	return &investapi.Quotation{
		Units: d.IntPart(),
		Nano:  int32(nano),
	}
}

func TinToDec(units int64, nano int32) decimal.Decimal {
	whole := decimal.NewFromInt(units)
	fr := decimal.New(int64(nano), -9)
	return whole.Add(fr)
}

func DecToTin(d decimal.Decimal) (units int64, nano int32) {
	nano = int32(d.Mod(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(1_000_000_000)).IntPart())
	units = d.IntPart()
	return units, nano
}
