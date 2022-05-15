package repository

import (
	"gorm.io/gorm"
	"invest-robot/dto"
)

type StatRepository interface {
	GetAlgorithmStat(req *dto.StatAlgoRequest) (*dto.StatAlgoResponse, error)
}

type PgStatRepository struct {
	db *gorm.DB
}

func NewStatRepository(db *gorm.DB) StatRepository {
	return &PgStatRepository{db: db}
}

func (r *PgStatRepository) GetAlgorithmStat(req *dto.StatAlgoRequest) (*dto.StatAlgoResponse, error) {
	moneyStat := r.getMoneyStat(req.AlgorithmID)
	var successOp uint
	for _, st := range moneyStat {
		successOp += st.OperationNum
	}

	failedCountSql := "select count(*) from actions where algorithm_id = ? and status = 'FAILED'"
	var failedOp uint
	r.db.Raw(failedCountSql, req.AlgorithmID).Scan(&failedOp)

	instrStat := r.getInstrStat(req.AlgorithmID)

	res := &dto.StatAlgoResponse{
		AlgorithmID:       req.AlgorithmID,
		SuccessOrders:     successOp,
		FailedOrders:      failedOp,
		MoneyChanges:      moneyStat,
		InstrumentChanges: instrStat,
	}
	return res, nil
}

func (r *PgStatRepository) getMoneyStat(algoId uint) []dto.MoneyStat {
	moneyStatSql := `with sc as (select id,
                   currency,
                   updated_at,
                   case when direction = 0 then -amount else amount end as sign_amount
            from actions a
            where a.algorithm_id = ?
              and a.status = 'SUCCESS'),
     sc_wdw as (select currency,
                       row_number() over (partition by currency order by updated_at desc) as row,
                       count(id) over w                                                   as cnt,
                       sum(sign_amount) over w                                            as total_sum
                from sc
                    window w as (partition by currency))
     select currency, cnt as operation_num, total_sum as final_value
     from sc_wdw
     where row = 1`
	var moneyStat []dto.MoneyStat
	r.db.Raw(moneyStatSql, algoId).Scan(&moneyStat)
	return moneyStat
}

func (r *PgStatRepository) getInstrStat(algoId uint) []dto.InstrumentStat {
	instrumentStatSql := `with sc as (select id,
                   instr_figi,
                   amount / lot_amount                                          as lot_price,
                   case when direction = 0 then lot_amount else -lot_amount end as sign_lot_amount,
                   currency,
                   updated_at,
                   case when direction = 0 then -amount else amount end         as sign_amount
            from actions a
            where a.algorithm_id = ?
              and a.status = 'SUCCESS'),
     sc_wdw as (select instr_figi,
                       sum(sign_lot_amount) over ip                                         as lot_sum,
                       count(id) over ip                                                    as cnt,
                       lot_price,
                       row_number() over (partition by instr_figi order by updated_at desc) as row,
                       sum(sign_amount) over ip                                             as money_sum,
                       currency
                from sc
                    window ip as (partition by instr_figi))
     select instr_figi,
                       lot_sum   as final_amount,
                       cnt       as operation_num,
                       lot_price as last_lot_price,
                       money_sum as final_money_val,
                       currency
     from sc_wdw
     where row = 1`

	var instrStat []dto.InstrumentStat
	r.db.Raw(instrumentStatSql, algoId).Scan(&instrStat)
	return instrStat
}
