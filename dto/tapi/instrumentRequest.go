package tapi

import investapi "invest-robot/tapigen"

type InstrumentIdType int

const (
	InstrumentIdUnspecified InstrumentIdType = iota
	InstrumentIdTypeFigi
	InstrumentIdTypeTicker
	InstrumentIdTypeUID
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
