package trade

import (
	"context"
	"go.uber.org/zap"
	"invest-robot/collections"
	"invest-robot/domain"
	"invest-robot/repository"
	"invest-robot/service"
	"invest-robot/strategy/stmodel"
)

type ProdApiTrader struct {
	*BaseTrader
}

func (t *ProdApiTrader) Go(ctx context.Context) {
	t.ctx = ctx
	go t.checkOrdersBg()
	go t.actionProcBg()
}

func NewProdTrader(infoSrv service.InfoSrv, tradeSrv service.TradeService, actionRep repository.ActionRepository, logger *zap.SugaredLogger) Trader {
	return &ProdApiTrader{
		&BaseTrader{
			infoSrv:   infoSrv,
			tradeSrv:  tradeSrv,
			actionRep: actionRep,
			subs:      collections.NewSyncMap[uint, *stmodel.Subscription](),
			orders:    collections.NewSyncMap[string, *domain.Action](),
			algoCh:    make(chan *stmodel.ActionReq, 1),
			logger:    logger,
		},
	}
}
