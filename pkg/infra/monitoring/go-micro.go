package monitoring

import (
	microprometheus "github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/v3/server"
)

func GoMicroHandlerWrapper() server.HandlerWrapper {
	return microprometheus.NewHandlerWrapper(func(opts *microprometheus.Options) {
		opts.Name = defaultConf.ServerName
	})
}
