package domain

import (
	"github.com/shopspring/decimal"
	"time"
)

type ActionStatus string
type ActionDirection int

const (
	CREATED  ActionStatus = "CREATED"
	POSTED   ActionStatus = "POSTED"
	CANCELED ActionStatus = "CANCELED"
	SUCCESS  ActionStatus = "SUCCESS"
	FAILED   ActionStatus = "FAILED"
)

const (
	BUY ActionDirection = iota
	SELL
)

type Action struct {
	ID            uint
	AlgorithmID   uint
	AccountID     string
	Direction     ActionDirection
	InstrFigi     string //instrument figi to buy/sell
	LotAmount     int64  //amount of instrument to sell in case of sell
	Status        ActionStatus
	Info          string          //failed details etc.
	Currency      string          //real currency used for buy/sell
	Amount        decimal.Decimal //real amount with taxes
	PositionPrice decimal.Decimal //position price
	OrderId       string
	RetrievedAt   time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
