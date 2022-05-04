package avr

import (
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
)

type ProdDataProc struct {
	algo    *domain.Algorithm
	infoSrv service.InfoSrv
	hRep    repository.HistoryRepository
}

func (a *ProdDataProc) GetDataStream() (<-chan procData, error) {
	return nil, errors.NewNotImplemented()
}

func (a *ProdDataProc) Go() {
	//todo implement me
}

func (a *ProdDataProc) Stop() error {
	return errors.NewNotImplemented()
}

func newProdDataProc(req *domain.Algorithm, hRep repository.HistoryRepository, infoSrv service.InfoSrv) (DataProc, error) {
	return &ProdDataProc{algo: req, hRep: hRep, infoSrv: infoSrv}, nil
}
