package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wythers/algalon/engine"
	"github.com/wythers/algalon/pipeline"
	"github.com/wythers/algalon/utils"
)

func Suspend(g *engine.Engine) gin.HandlerFunc {
	type request struct {
		Apps []string `json:"apps" binding:"required"`
	}

	type response struct {
		Suspended []string `json:"suspended"`
	}

	return gin.HandlerFunc(func(c *gin.Context) {

		if g.Isclosed() {
			c.AbortWithError(http.StatusServiceUnavailable, utils.ErrClosed)
			return
		}

		var (
			r    request
			resp response
		)

		if err := c.BindJSON(&r); err != nil {
			c.AbortWithError(http.StatusBadRequest, utils.ErrAppKeyInvalid)
			return
		}

		info, err := g.IsCerted(r.Apps)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, utils.ErrAppKeyInvalid)
			return
		}

		p := pipeline.New(g)
		for _, app := range r.Apps {
			ws, err := g.Wallets.Suspend(app)
			if err != nil {
				continue
			}

			for _, w := range ws {
				p.Pipe(w, 1<<2, info[app])
			}

			resp.Suspended = append(resp.Suspended, app)
		}

		c.JSON(http.StatusOK, resp)
	})
}
