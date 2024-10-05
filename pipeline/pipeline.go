package pipeline

import (
	"sync"

	"github.com/shopspring/decimal"
	"github.com/wythers/algalon/engine"
	"github.com/wythers/algalon/utils"
	"github.com/wythers/algalon/wallets"
)

const (
	opCheck  = 1
	opFilter = 1 << 1
	opPump   = 1 << 2
)

type PipeLine struct {
	g *engine.Engine
}

var (
	Flow = CoinFlow{
		flow: make(map[string]struct {
			Inflow  decimal.Decimal
			Outflow decimal.Decimal
		}),
	}
)

func New(e *engine.Engine) PipeLine {
	return PipeLine{
		g: e,
	}
}

func (p PipeLine) Pipe(w *wallets.TRCWallet, opCode int, info *utils.AppInfo) (string, error) {
	var (
		diff       = decimal.Zero
		err  error = nil
	)

	if (opCode & opCheck) != 0 {
		diff, err = check(p.g, w, info)
	}

	half := func() {
		if err != nil {
			opCode = opPump
		}

		if (opCode & opFilter) != 0 {
			opCode = opCode | filter(p.g, w, info, diff)
		}

		if (opCode & opPump) != 0 {
			pump(p.g, w, info)
		}
	}

	select {
	case p.g.Ch <- half:
	default:
		go half()
	}

	return diff.String(), err
}

type CoinFlow struct {
	lock sync.RWMutex

	flow map[string]struct {
		Inflow  decimal.Decimal
		Outflow decimal.Decimal
	}
}

func (f *CoinFlow) GetFlow(k string) (string, string) {
	defer f.lock.RUnlock()
	f.lock.RLock()

	var (
		in  = f.flow[k].Inflow.String()
		out = f.flow[k].Outflow.String()
	)

	return in, out
}

func (f *CoinFlow) In(k string, diff decimal.Decimal) {
	defer f.lock.Unlock()
	f.lock.Lock()

	d := f.flow[k].Inflow.Add(diff)
	newf := struct {
		Inflow  decimal.Decimal
		Outflow decimal.Decimal
	}{
		Inflow:  d,
		Outflow: f.flow[k].Outflow,
	}

	f.flow[k] = newf
}

func (f *CoinFlow) Out(k string, diff decimal.Decimal) {
	defer f.lock.Unlock()
	f.lock.Lock()

	d := f.flow[k].Outflow.Add(diff)
	newf := struct {
		Inflow  decimal.Decimal
		Outflow decimal.Decimal
	}{
		Inflow:  f.flow[k].Inflow,
		Outflow: d,
	}

	f.flow[k] = newf
}
