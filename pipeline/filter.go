package pipeline

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wythers/algalon/engine"
	"github.com/wythers/algalon/utils"
	"github.com/wythers/algalon/wallets"
)

func filter(g *engine.Engine, w *wallets.TRCWallet, info *utils.AppInfo, diff decimal.Decimal) int {
	var (
		threshold, _ = decimal.NewFromString(info.Threshold)
	)

	if !diff.IsZero() {
		Flow.In(w.Owner, diff)

		r := utils.Record{
			Owner:     w.Owner,
			Amount:    diff.String(),
			Type:      "inflow",
			Timestamp: time.Now().Unix(),
		}

		err := g.W.Write(r)
		if err != nil {
			j, _ := json.Marshal(r)
			g.L.Println(string(j))
		}
	}

	if threshold.LessThanOrEqual(w.Trx) {
		return opPump
	}

	err := g.Rel(w)
	if err == nil {
		return 0
	}

	return opPump
}
