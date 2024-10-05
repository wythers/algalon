package wallets

import (
	"errors"
	"sync"

	"github.com/wythers/algalon/utils"
)

type Track utils.Queuer[TRCWallet]

type TRCWallets struct {
	RWlock sync.RWMutex

	Spec   map[string]Track
	Normal Track
}

func (t *TRCWallets) Track(key string) {
	defer t.RWlock.Unlock()
	t.RWlock.Lock()

	if r := t.Spec[key]; r != nil {
		return
	}

	t.Spec[key] = utils.NewQueue[TRCWallet]()
}

func (t *TRCWallets) Count(key string) int {
	defer t.RWlock.RUnlock()
	t.RWlock.RLock()

	if key == "" {
		return t.Normal.Counter()
	}

	track, ok := t.Spec[key]
	if !ok {
		return 0
	}

	return track.Counter()
}

func (t *TRCWallets) Suspend(key string) ([]*TRCWallet, error) {
	defer t.RWlock.RUnlock()
	t.RWlock.RLock()

	track, ok := t.Spec[key]
	if !ok {
		return nil, errors.New("algalon: unknown key")
	}

	return track.Suspend()
}

func (t *TRCWallets) Close() []*TRCWallet {
	defer t.RWlock.RUnlock()
	t.RWlock.RLock()

	var all []*TRCWallet
	for _, track := range t.Spec {
		w, err := track.Suspend()
		if err != nil {
			continue
		}

		all = append(all, w...)
	}

	return all
}
