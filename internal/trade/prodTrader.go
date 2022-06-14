package trade

import (
	"context"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/collections"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/entity"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/repository"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/service"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/strategy/stmodel"
	"go.uber.org/zap"
)

type ProdTrader struct {
	*BaseTrader
}

func (t *ProdTrader) Go(ctx context.Context) {
	t.ctx = ctx
	go t.checkOrdersBg()
	go t.actionProcBg()
}

func NewProdTrader(infoSrv service.InfoSrv, tradeSrv service.TradeService, actionRep repository.ActionRepository, logger *zap.SugaredLogger) Trader {
	return &ProdTrader{
		&BaseTrader{
			infoSrv:   infoSrv,
			tradeSrv:  tradeSrv,
			actionRep: actionRep,
			subs:      collections.NewSyncMap[uint, *stmodel.Subscription](),
			orders:    collections.NewSyncMap[string, *entity.Action](),
			algoCh:    make(chan *stmodel.ActionReq, 1),
			logger:    logger,
		},
	}
}
