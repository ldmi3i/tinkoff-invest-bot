package dtotapi

type SubscribeCandlesRequest struct {
	Instruments []Instrument
}

type Instrument struct {
	Figi string
	//0 Undefined
	//1 Min interval
	//2 Five min interval
	Interval int
}
