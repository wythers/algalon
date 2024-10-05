package algalon

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wythers/algalon/engine"
	"github.com/wythers/algalon/handlers"
	"github.com/wythers/algalon/pipeline"
	"github.com/wythers/algalon/utils"
)

type Writer = utils.Writer
type Record = utils.Record

type Algalon struct {
	g *engine.Engine

	c *gin.Engine
}

func New(w Writer, opts ...gin.OptionFunc) Algalon {
	g := engine.New(w)
	c := gin.New(opts...)
	setRouter(c, g)

	return Algalon{g: g, c: c}
}

func Default(w Writer) Algalon {
	g := engine.New(w)
	c := gin.Default()
	setRouter(c, g)

	return Algalon{g: g, c: c}
}

func (al Algalon) Run(adr ...string) {
	c := al.c
	g := al.g

	address := resolveAddress(adr)
	server := &http.Server{
		Addr:    address,
		Handler: c,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		stop()

		log.Println("[algalon]: algalon starts closing...")

		var (
			tmp []string
			n   int64
		)

		for {
			n = atomic.LoadInt64(&g.Count)
			if n == -1 {
				break
			}

			success := atomic.CompareAndSwapInt64(&g.Count, n, -1)
			if success {
				break
			}
		}

		if n != -1 {
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
		}
		log.Println("[algalon]: algalon has been closed gracefully.")

		_ = server.Close()
	}()

	if err := server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Println("[algalon]: gin server closed under request")
		} else {
			log.Fatalf("[algalon]: gin server closed unexpect: %s\n", err.Error())
		}
	}
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			fmt.Printf("[algalon]: environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		fmt.Println("[algalon]: environment variable PORT is undefined. using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("[algalon]: too many parameters")
	}
}

func setRouter(c *gin.Engine, g *engine.Engine) {
	// normal APIs
	{
		c.GET("/wallet", handlers.Wallet(g))
		c.GET("/query", handlers.Query(g))
	}

	// admin APIs
	{
		c.GET("/matrix", handlers.Matrix(g))
		c.POST("/suspend", handlers.Suspend(g))
		c.POST("/update", handlers.Update(g))
		c.POST("/close", handlers.Close(g))
	}
}
