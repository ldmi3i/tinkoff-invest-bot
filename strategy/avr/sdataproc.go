package avr

import (
	"invest-robot/domain"
	"invest-robot/errors"
	"invest-robot/repository"
	"invest-robot/service"
)

type SandboxDataProc struct {
	algo    *domain.Algorithm
	infoSrv service.InfoSrv
	hRep    repository.HistoryRepository
}

func (a *SandboxDataProc) GetDataStream() (<-chan procData, error) {
	return nil, errors.NewNotImplemented()
}

func (a *SandboxDataProc) Go() {
	//todo implement me
}

func (a *SandboxDataProc) Stop() error {
	return errors.NewNotImplemented()
}

func newSandboxDataProc(req *domain.Algorithm, hRep repository.HistoryRepository, infoSrv service.InfoSrv) (DataProc, error) {
	return &SandboxDataProc{algo: req, infoSrv: infoSrv, hRep: hRep}, nil
}
