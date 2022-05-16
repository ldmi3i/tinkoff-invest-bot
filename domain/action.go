package domain

import (
	"github.com/shopspring/decimal"
	"time"
)

type ActionStatus string
type ActionDirection int
type OrderType string

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

const (
	LIMITED OrderType = "LIMITED"
	MARKET  OrderType = "MARKET"
)

type Action struct {
	ID             uint
	AlgorithmID    uint
	AccountID      string
	Direction      ActionDirection
	InstrFigi      string    //instrument figi to buy/sell
	LotAmount      int64     //amount of instrument to sell in case of sell
	OrderType      OrderType //type of order - is it limited request or market
	Status         ActionStatus
	Info           string          //failed details etc.
	Currency       string          //real currency used for buy/sell
	TotalPrice     decimal.Decimal `gorm:"type:numeric"` //real full amount with taxes
	ReqPrice       decimal.Decimal `gorm:"type:numeric"` //for limited order requested amount of currency
	PositionPrice  decimal.Decimal `gorm:"type:numeric"` //position price
	LotsExecuted   int64           `gorm:"default:0"`
	ExpirationTime time.Time       //Expiration time of order - if order is Partially filled and expired - cancel will be sent
	OrderId        string
	RetrievedAt    time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
