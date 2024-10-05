package engine

import (
	"log"
	"os"
	"sync"
	"sync/atomic"

	"golang.org/x/sys/cpu"

	"github.com/shopspring/decimal"
	"github.com/wythers/algalon/cert"
	"github.com/wythers/algalon/provider"
	"github.com/wythers/algalon/provider/tron"
	"github.com/wythers/algalon/utils"
	"github.com/wythers/algalon/wallets"
)

type TaskType = func()

var (
	defaultTimeout  = 2
	defaultInterval = 1
	defaultTries    = 1
	defaultChSize   = 1024
	defaultWorks    = 8
)

const (
	fullNodeUrl     string = "https://api.trongrid.io"
	solidityNodeURL string = "https://api.trongrid.io"

	usdtContractAddress string = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
)

type Engine struct {
	// St means static table
	St StaticTable

	Wallets wallets.TRCWallets

	Provider provider.Providerer

	W utils.Writer
	L *log.Logger

	Certs cert.Cert

	// seconds
	Timeout int
	// seconds
	Interval int
	Tries    int

	Ch chan TaskType

	_ cpu.CacheLinePad

	ExcepCount int64

	_ cpu.CacheLinePad

	// count and closed togather to occupy one cache line
	Count  int64
	Closed int64
}

func New(w utils.Writer) *Engine {
	g := &Engine{
		St: StaticTable{
			Table: make(map[string]*wallets.TRCWallet),
		},
		Wallets: wallets.TRCWallets{
			Spec:   make(map[string]wallets.Track),
			Normal: utils.NewQueue[wallets.TRCWallet](),
		},

		Provider: tron.New(fullNodeUrl, solidityNodeURL, usdtContractAddress),
		W:        w,
		L:        log.New(os.Stdout, "algalon: ", log.LstdFlags|log.Lmsgprefix),

		Certs: cert.Cert{
			Certs: make(map[string]utils.AppInfo),
		},

		Timeout:  defaultTimeout,
		Interval: defaultInterval,
		Tries:    defaultTries,

		Ch: make(chan TaskType, defaultChSize),

		Count:  0,
		Closed: 0,
	}

	for i := 0; i < defaultWorks; i++ {
		go func() {
			for task := range g.Ch {
				task()
			}
		}()
	}

	return g
}

type StaticTable struct {
	RWlock sync.RWMutex

	Table map[string]*wallets.TRCWallet
}

func (s *StaticTable) Push(address string, w *wallets.TRCWallet) {
	defer s.RWlock.Unlock()
	s.RWlock.Lock()

	s.Table[address] = w
}

func (s *StaticTable) Map(address string) *wallets.TRCWallet {
	defer s.RWlock.RUnlock()
	s.RWlock.RLock()

	return s.Table[address]
}

func (e *Engine) Isclosed() bool {
	return atomic.LoadInt64(&e.Count) == -1
}

func (e *Engine) IsCerted(key []string) (map[string]*utils.AppInfo, error) {
	return e.Certs.IsCerted(key)
}

func (e *Engine) Alloc(key string) (*wallets.TRCWallet, error) {
	defer e.Wallets.RWlock.RUnlock()
	e.Wallets.RWlock.RLock()

	// a valid key matches a track, so there is no need to check whether the track exists
	track := e.Wallets.Spec[key]
	w, err := track.Dequeue()
	if err != nil && err != utils.ErrEOF {
		return nil, err
	}

	if w != nil {
		return w, nil
	}

	w, err = e.Wallets.Normal.Dequeue()
	if err == nil && w != nil {
		w.Owner = key
		return w, nil
	}

	raw, err := e.Provider.GenerateWallet()
	if err != nil {
		return nil, err
	}

	for {
		tmp := atomic.LoadInt64(&e.Count)
		if tmp == -1 {
			return nil, utils.ErrClosed
		}

		if ok := atomic.CompareAndSwapInt64(&e.Count, tmp, tmp+1); ok {
			break
		}
	}

	w = &wallets.TRCWallet{
		W:     raw,
		Owner: key,
		Sqno:  1,

		Trx:   decimal.Zero,
		Trc20: decimal.Zero,
	}

	e.St.Push(w.W.Address, w)

	return w, nil
}

// means release, or free
func (e *Engine) Rel(w *wallets.TRCWallet) error {
	var (
		owner = w.Owner
	)

	defer e.Wallets.RWlock.RUnlock()
	e.Wallets.RWlock.RLock()

	if owner == "" {
		return e.Wallets.Normal.Enqueue(w)
	}

	track := e.Wallets.Spec[owner]
	return track.Enqueue(w)
}
