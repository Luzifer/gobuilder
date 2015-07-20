package main

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricActiveWorkers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gobuilder_workers",
		Help: "Current count of active workers",
	})
	metricQueueLength = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gobuilder_queuelength",
		Help: "Current number of items in the build queue",
	})
)

func init() {
	prometheus.MustRegister(metricActiveWorkers)
	prometheus.MustRegister(metricQueueLength)

	go fetchMetrics()
}

func fetchMetrics() {
	for {
		// Fetch clients active in last 5min
		timestamp := strconv.Itoa(int(time.Now().Unix() - 300))
		activeWorkers, _ := redisClient.ZCount("active-workers", timestamp, "+inf")
		metricActiveWorkers.Set(float64(activeWorkers))

		queueLength, _ := redisClient.LLen("build-queue")
		metricQueueLength.Set(float64(queueLength))

		<-time.After(30 * time.Second)
	}
}
