package pipeline

import (
	"encoding/json"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wythers/algalon/engine"
	"github.com/wythers/algalon/utils"
	"github.com/wythers/algalon/wallets"
)

var (
	noise, _ = decimal.NewFromString("0.1")
)

func pump(g *engine.Engine, w *wallets.TRCWallet, info *utils.AppInfo) {
	var (
		to_address = info.Address[rand.IntN(len(info.Address))]
		tronApiKey = info.TronApiKey
		amount     = w.Trx
	)

	if amount.LessThan(noise) {
		w.Owner = ""
		g.Rel(w)
		return
	}

	Txid, err := g.Provider.SendTRX(w.W, to_address, amount, tronApiKey)
	r := utils.Record{
		From: w.W.Address,
		To:   to_address,

		Amount: amount.String(),

		TxId: Txid,

		Timestamp: time.Now().Unix(),
		Owner:     w.Owner,
	}

	if err != nil {
		r.PrivKey = w.W.PrivKey
		r.Type = "exception"

		atomic.AddInt64(&g.ExcepCount, 1)
	} else {
		Flow.Out(w.Owner, w.Trx)

		r.Type = "outflow"
		w.Trx = decimal.Zero
		w.Owner = ""

		g.Rel(w)
	}

	err = g.W.Write(r)
	if err != nil {
		j, _ := json.Marshal(r)
		g.L.Println(string(j))
	}
}
