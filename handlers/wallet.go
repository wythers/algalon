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

func Wallet(g *engine.Engine) gin.HandlerFunc {
	type request struct {
		AlgalonAppKey string `json:"algalon_app_key" binding:"required"`
	}

	type response struct {
		Address string `json:"address"`
		Sqno    uint64 `json:"sqno"`
	}

	return gin.HandlerFunc(func(c *gin.Context) {

		if g.Isclosed() {
			c.AbortWithError(http.StatusServiceUnavailable, utils.ErrClosed)
			return
		}

		var (
			r    request
			info *utils.AppInfo
		)

		if err := c.BindJSON(&r); err != nil {
			c.AbortWithError(http.StatusBadRequest, utils.ErrAppKeyInvalid)
			return
		}

		var key []string
		key = append(key, r.AlgalonAppKey)
		tmp, err := g.IsCerted(key)

		if err != nil {
			c.AbortWithError(http.StatusBadRequest, utils.ErrAppKeyInvalid)
			return
		}
		info = tmp[r.AlgalonAppKey]

		w, err := g.Alloc(r.AlgalonAppKey)
		if err != nil {
			c.AbortWithError(http.StatusForbidden, err)
			return
		}

		sqno := atomic.LoadUint64(&w.Sqno)

		time.AfterFunc(time.Duration(g.Timeout)*time.Second, func() {
			success := atomic.CompareAndSwapUint64(&w.Sqno, sqno, sqno+1)
			if !success {
				return
			}

			// create a pipeline
			p := pipeline.New(g)
			p.Pipe(w, 1<<1, info)
		})

		resp := response{
			Address: w.W.Address,
			Sqno:    sqno,
		}

		c.JSON(http.StatusOK, resp)
	})
}
