package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"template/internal/client/rpc"
	"template/pkg/proto"
)

var (
	consulAddr string
)

func init() {
	flag.StringVar(&consulAddr, "consul", "127.0.0.1:8500", "consul address")
	flag.Parse()
}

func main() {
	cli := rpc.NewGreeterClient(consulAddr)
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		rsp, err := cli.Hello(context.TODO(), &proto.HelloRequest{
			Name: "libz",
		})

		if err != nil {
			log.Println(err.Error())
			c.AbortWithStatus(http.StatusGatewayTimeout)
			return
		}

		c.String(http.StatusOK, rsp.String())
	})

	pprof.Register(r)
	_ = r.Run(":28086")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch
}
