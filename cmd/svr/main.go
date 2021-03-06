package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"template/internal/app"
)

var version = "No Build Info"

func main() {
	srv, err := app.New(
		app.Config(),
		app.Version(version),
		app.NodeID(),
		app.Logger(),
		app.KVStore(),
		app.Monitor(),
		app.RedisCli(),
		app.MySQLCli(),
		app.MongoCli(),
		app.Dao(),
		app.UseCase(),
		app.WebService(),
		app.RpcService(),
	)

	defer func() {
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		}
	}()

	if err != nil {
		return
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	err = srv.Run(ch)
	<-ch
}
