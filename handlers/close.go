package handlers

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wythers/algalon/engine"
	"github.com/wythers/algalon/pipeline"
	"github.com/wythers/algalon/utils"
)

func Close(g *engine.Engine) gin.HandlerFunc {
	type response struct {
		Status string `json:"Status"`
	}

	return gin.HandlerFunc(func(c *gin.Context) {

		if g.Isclosed() {
			c.AbortWithError(http.StatusServiceUnavailable, utils.ErrClosed)
			return
		}

		var (
			tmp []string
			n   int64
		)

		// lock the total number of wallets
		for {
			n = atomic.LoadInt64(&g.Count)
			if n == -1 {
				c.AbortWithError(http.StatusServiceUnavailable, utils.ErrClosed)
				return
			}

			success := atomic.CompareAndSwapInt64(&g.Count, n, -1)
			if success {
				break
			}
		}

		//		n := atomic.SwapInt64(&g.Count, -1)
		atomic.StoreInt64(&g.Closed, n)
		p := pipeline.New(g)
		apps, _ := g.Certs.IsCerted(tmp)

		all := g.Wallets.Close()

		for _, w := range all {
			p.Pipe(w, 1<<2, apps[w.Owner])
		}

		for {
			total := int64(g.Wallets.Count("")) + atomic.LoadInt64(&g.ExcepCount)

			if total == n {
				break
			}

			time.Sleep(2 * time.Second)
		}

		c.JSON(http.StatusOK, response{Status: "done"})
	})
}
