package pipeline

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wythers/algalon/engine"
	"github.com/wythers/algalon/utils"
	"github.com/wythers/algalon/wallets"
)

func check(g *engine.Engine, w *wallets.TRCWallet, info *utils.AppInfo) (decimal.Decimal, error) {
	var (
		tries  = g.Tries
		diff   = decimal.Zero
		min, _ = decimal.NewFromString(info.Mta)
	)

	for {
		amount, err := g.Provider.GetTRXBalance(w.W.Address, info.TronApiKey)
		if err != nil {
			return decimal.Decimal{}, fmt.Errorf("algalon: %s: %s", w.W.Address, err.Error())
		}

		diff = diff.Add(amount.Sub(w.Trx))
		w.Trx = amount
		if min.LessThanOrEqual(diff) {
			break
		}

		tries -= 1

		if tries == 0 {
			break
		}

		time.Sleep(time.Duration(g.Interval) * time.Second)
	}

	return diff, nil
}
