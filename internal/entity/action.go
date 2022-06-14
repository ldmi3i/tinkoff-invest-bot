package entity

import (
	"github.com/shopspring/decimal"
	"time"
)

type ActionStatus string
type ActionDirection int
type OrderType string

const (
	Created  ActionStatus = "CREATED"
	Posted   ActionStatus = "POSTED"
	Canceled ActionStatus = "CANCELED"
	Success  ActionStatus = "SUCCESS"
	Failed   ActionStatus = "FAILED"
)

const (
	Buy ActionDirection = iota
	Sell
)

const (
	Limited OrderType = "LIMITED"
	Market  OrderType = "MARKET"
)

//Action represents data used in all trade order steps. Algorithm initiate action request to trader.
//Some data must be populated from algorithm, other filled by trader
//each field has a mark when it must be filled
type Action struct {
	ID             uint            //Filled on save;
	AlgorithmID    uint            //Filled on init;
	AccountID      string          //Filled on init;
	Direction      ActionDirection //Filled by algorithm on request; Direction of operation buy/sell
	InstrFigi      string          //Filled by algorithm (required); instrument figi to buy/sell
	LotAmount      int64           //Filled by algorithm (optional); amount of instrument to sell in case of sell
	OrderType      OrderType       //Filled by algorithm (required); type of order - is it limited request or market
	Status         ActionStatus    //Filled by algorithm as Created, then trader update it status; Represents current order trade status
	Info           string          //Filled by trader; failed details etc.
	Currency       string          //Filled by algorithm (required); real currency used for buy/sell
	TotalPrice     decimal.Decimal `gorm:"type:numeric"` //Filled by trader; real full amount with taxes
	ReqPrice       decimal.Decimal `gorm:"type:numeric"` //Filled by algorithm (optionally); Price of position for limited order
	PositionPrice  decimal.Decimal `gorm:"type:numeric"` //Filled by trader; Average position price returned from Tinkoff API - may be used by algorithm
	LotsExecuted   int64           `gorm:"default:0"`    //Filled by trader; Number of lots executed
	ExpirationTime time.Time       //Filled by algorithm (optional); Expiration time of order - if order is Partially filled and expired - cancel will be sent
	OrderId        string          //Filled by trader; Order id returned by Tinkoff API
	RetrievedAt    time.Time       //Filled by data processor (if any); Time when data was retrieved from API
	CreatedAt      time.Time       //Filled by gorm on insert
	UpdatedAt      time.Time       //Filled by gorm on update
}
