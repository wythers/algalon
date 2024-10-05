package handlers

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/wythers/algalon/engine"
	"github.com/wythers/algalon/pipeline"
	"github.com/wythers/algalon/utils"
)

func Query(g *engine.Engine) gin.HandlerFunc {
	type request struct {
		AlgalonAppKey string `json:"algalon_app_key" binding:"required"`
		Address       string `json:"address" binding:"required"`
		Sqno          uint64 `json:"sqno" binding:"required"`
	}

	type response struct {
		Address string `json:"address"`
		Amount  string `json:"amount,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	return gin.HandlerFunc(func(c *gin.Context) {
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

		sqno := r.Sqno
		w := g.St.Map(r.Address)
		if w == nil {
			c.AbortWithError(http.StatusBadRequest, utils.ErrAppKeyInvalid)
			return
		}

		success := atomic.CompareAndSwapUint64(&w.Sqno, uint64(sqno), uint64(sqno)+1)
		if !success {
			c.AbortWithError(http.StatusRequestTimeout, utils.ErrTimeOut)
			return
		}

		p := pipeline.New(g)
		amount, err := p.Pipe(w, 1|1<<1, info)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response{
				Address: r.Address,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, response{
			Address: r.Address,
			Amount:  amount,
		})
	})
}
