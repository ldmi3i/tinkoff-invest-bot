package tapi

import investapi "invest-robot/tapigen"

type InstrumentIdType int

const (
	INSTRUMENT_ID_UNSPECIFIED InstrumentIdType = iota
	INSTRUMENT_ID_TYPE_FIGI
	INSTRUMENT_ID_TYPE_TICKER
	INSTRUMENT_ID_TYPE_UID
)

type InstrumentRequest struct {
	IdType    InstrumentIdType
	ClassCode string
	Id        string
}

func (req *InstrumentRequest) ToTinApi() *investapi.InstrumentRequest {
	return &investapi.InstrumentRequest{
		IdType:    investapi.InstrumentIdType(req.IdType),
		ClassCode: req.ClassCode,
		Id:        req.Id,
	}
}
