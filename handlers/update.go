package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wythers/algalon/engine"
	"github.com/wythers/algalon/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func Update(g *engine.Engine) gin.HandlerFunc {
	type app struct {
		AlgalonAppKey string `json:"algalon_app_key,omitempty"`

		Threshold string `json:"threshold" binding:"required"`
		// Minimum transaction amount
		Mta string `json:"mta" binding:"required"`

		Address    []string `json:"address" binding:"required"`
		TronApiKey []string `json:"tron_api_key" binding:"required"`
	}

	type request struct {
		Apps []app `json:"apps"`
	}

	type response struct {
		AlgalonAppKey []string `json:"algalon_app_key"`
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

		for i := 0; i < len(r.Apps); i++ {
			var k string
			if r.Apps[i].AlgalonAppKey == "" {
				b, _ := bson.NewObjectID().MarshalText()
				k = string(b)
			} else {
				k = r.Apps[i].AlgalonAppKey
			}

			g.Wallets.Track(k)

			info := utils.AppInfo{
				Threshold:  r.Apps[i].Threshold,
				Mta:        r.Apps[i].Mta,
				Address:    r.Apps[i].Address,
				TronApiKey: r.Apps[i].TronApiKey,
			}

			g.Certs.Add(k, info)
			resp.AlgalonAppKey = append(resp.AlgalonAppKey, k)
		}

		c.JSON(http.StatusOK, resp)
	})
}
