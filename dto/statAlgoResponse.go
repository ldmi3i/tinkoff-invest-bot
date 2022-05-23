package dto

import "github.com/shopspring/decimal"

type StatAlgoResponse struct {
	AlgorithmID       uint
	SuccessOrders     uint             //Number of success orders
	FailedOrders      uint             //Number of failed orders
	CanceledOrders    uint             //Number of cancelled orders
	MoneyChanges      []MoneyStat      //Data about how money amount changed by each currency
	InstrumentChanges []InstrumentStat //Data about how instrument amount changed by each instrument
}

type MoneyStat struct {
	Currency     string          //Currency of money
	FinalValue   decimal.Decimal //Final value of currency balance (starting value as 0)
	OperationNum uint            //Number of success operations with currency
}

type InstrumentStat struct {
	InstrFigi     string          //Instrument figi
	FinalAmount   int64           //Final amount of lots of instrument
	OperationNum  uint            //Number of success operations with instrument made by algorithm
	LastLotPrice  decimal.Decimal //Last price of instrument
	FinalMoneyVal decimal.Decimal //Final money balance by operations with instrument (sum of money spend/receive by instrument figi)
	Currency      string          //Currency of instrument
}
