package cert

import (
	"sync"

	"github.com/wythers/algalon/utils"
)

type Cert struct {
	RWlock sync.RWMutex

	Certs map[string]utils.AppInfo
}

func (c *Cert) IsCerted(key []string) (map[string]*utils.AppInfo, error) {
	defer c.RWlock.RUnlock()
	c.RWlock.RLock()

	var (
		m = make(map[string]*utils.AppInfo)
	)

	if len(key) == 0 {
		for k, i := range c.Certs {
			m[k] = &i
		}

		return m, nil
	}

	for _, k := range key {
		i, ok := c.Certs[k]
		if !ok {
			return nil, utils.ErrAppKeyInvalid
		}

		m[k] = &i
	}

	return m, nil
}

func (c *Cert) Add(key string, info utils.AppInfo) {
	defer c.RWlock.Unlock()
	c.RWlock.Lock()

	c.Certs[key] = info
}
