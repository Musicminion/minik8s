package main

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

var bgCtx = context.Background()
var cancel context.CancelFunc

func consumeHigherCpu(ctx context.Context) {
	for i := 0; i < 5; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
	}
}

func consumeHigherMemory(ctx context.Context) {
	for i := 0; i < 5; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// 申请更多的内存
					s := make([]byte, 1024*1024*200) // 200 MB
					_ = s
				}
			}
		}()
	}
}

func higherCpu(c *gin.Context) {
	if cancel != nil {
		return
	}

	var ctx context.Context
	ctx, cancel = context.WithCancel(bgCtx)
	consumeHigherCpu(ctx)
	c.String(http.StatusOK, "higher cpu utilization!")
}

func lowerCpu(c *gin.Context) {
	if cancel != nil {
		cancel()
		cancel = nil
	}
	c.String(http.StatusOK, "lower cpu utilization!")
}

func higherMemory(c *gin.Context) {
	if cancel != nil {
		return
	}

	var ctx context.Context
	ctx, cancel = context.WithCancel(bgCtx)
	consumeHigherMemory(ctx)
	c.String(http.StatusOK, "higher memory utilization!")
}

func lowerMemory(c *gin.Context) {
	if cancel != nil {
		cancel()
		cancel = nil
	}
	debug.FreeOSMemory()
	c.String(http.StatusOK, "lower memory utilization!")
}

func main() {
	r := gin.Default()
	r.GET("/hc", higherCpu)
	r.GET("/lc", lowerCpu)
	r.GET("/hm", higherMemory)
	r.GET("/lm", lowerMemory)

	log.Fatal(r.Run(":8090"))
}
