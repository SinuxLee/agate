package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	redisOpsCounter *prometheus.CounterVec
	redisHistogram  *prometheus.HistogramVec
)

func initRedis() {
	c := createCollector(defaultConf.ServerName, "redis", "command_count", "counter_vec", []string{"command", "instance", "status"})
	redisOpsCounter = c.(*prometheus.CounterVec)
	c = createCollector(defaultConf.ServerName, "redis", "command_duration_seconds", "histogram_vec", []string{"command", "instance", "status"})
	redisHistogram = c.(*prometheus.HistogramVec)
}

func GetRecordRedisCallStatsHandler(command, instance string) func(err error) {
	startTime := time.Now()

	return func(err error) {
		elapsed := float64(time.Since(startTime).Nanoseconds()) / 1e6
		status := "OK"
		if err != nil {
			status = "ERROR"
		}
		redisHistogram.WithLabelValues(command, instance, status).Observe(elapsed)
		redisOpsCounter.WithLabelValues(command, instance, status).Inc()
	}
}
