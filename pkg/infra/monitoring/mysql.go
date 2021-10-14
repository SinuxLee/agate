package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zsais/go-gin-prometheus"
)

var (
	mysqlOpsCounter *prometheus.CounterVec
	mysqlHistogram  *prometheus.HistogramVec
)

func initMysql() {
	c := createCollector(defaultConf.ServerName, "mysql", "client_calls", "counter_vec", []string{"method", "instance", "status"})
	mysqlOpsCounter = c.(*prometheus.CounterVec)
	c = createCollector(defaultConf.ServerName, "mysql", "client_duration_seconds", "histogram_vec", []string{"method", "instance", "status"})
	mysqlHistogram = c.(*prometheus.HistogramVec)
}

func GetRecordMysqlCallStatsHandler(method, instance string) func(err error) {
	startTime := time.Now()

	return func(err error) {
		elapsed := float64(time.Since(startTime).Nanoseconds()) / 1e6
		status := "OK"
		if err != nil {
			status = "ERROR"
		}
		mysqlHistogram.WithLabelValues(method, instance, status).Observe(elapsed)
		mysqlOpsCounter.WithLabelValues(method, instance, status).Inc()
	}
}

func createCollector(server, id, name, mtype string, args []string) prometheus.Collector {
	m := &ginprometheus.Metric{
		ID: id, Name: name, Type: mtype, Args: args,
	}

	metric := ginprometheus.NewMetric(m, server)

	if err := prometheus.Register(metric); err != nil {
		return nil
	}
	return metric
}
