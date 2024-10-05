package handlers

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/wythers/algalon/engine"
	"github.com/wythers/algalon/pipeline"
	"github.com/wythers/algalon/utils"
)

func Matrix(g *engine.Engine) gin.HandlerFunc {
	type request struct {
		Apps []string `json:"apps" binding:"required"`
	}

	type Appinfo struct {
		AlgalonAppKey string `json:"algalon_app_key"`

		State struct {
			Track   int    `json:"track"`
			Inflow  string `json:"inflow"`
			Outflow string `json:"outflow"`
		} `json:"state"`

		Config *utils.AppInfo `json:"config"`
	}

	type response struct {
		Algalon struct {
			State struct {
				Track           int   `json:"track"`
				Wallets         int64 `json:"wallets"`
				AbnormalWallets int64 `json:"abnormal_wallets"`
			} `json:"state"`
		}
		Apps []Appinfo `json:"apps"`
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		var (
			r    request
			resp response
		)

		if err := c.BindJSON(&r); err != nil {
			c.AbortWithError(http.StatusBadRequest, utils.ErrAppKeyInvalid)
			return
		}

		apps, err := g.IsCerted(r.Apps)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, utils.ErrAppKeyInvalid)
			return
		}

		for k, info := range apps {
			in, out := pipeline.Flow.GetFlow(k)
			cnt := g.Wallets.Count(k)

			var appInfo Appinfo

			appInfo.AlgalonAppKey = k
			appInfo.State.Inflow = in
			appInfo.State.Outflow = out
			appInfo.State.Track = cnt
			appInfo.Config = info

			resp.Apps = append(resp.Apps, appInfo)
		}

		all := atomic.LoadInt64(&g.Count)
		if all == -1 {
			all = atomic.LoadInt64(&g.Closed)
		}
		resp.Algalon.State.Track = g.Wallets.Count("")
		resp.Algalon.State.Wallets = all
		resp.Algalon.State.AbnormalWallets = atomic.LoadInt64(&g.ExcepCount)

		c.JSON(http.StatusOK, resp)
	})
}
