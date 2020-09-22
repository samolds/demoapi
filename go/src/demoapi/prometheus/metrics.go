package prometheus

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

//
// example usage of metrics in practice (without labels):
//
//     UsersGauge.Inc()
//     UsersGauge.Set(10)
//

var (
	// collectors used in the basic http request middleware
	httpRequestInFlightGauge = prom.NewGaugeVec(
		prom.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "A gauge active requests",
		}, []string{"service"})
	httpRequestCounter = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "http_requests_total",
			Help: "A counter of HTTP requests served",
		}, []string{"service", "code", "method"})
	httpRequestDuration = prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "http_requests_duration_seconds",
			Help:    "A histogram of HTTP request durations",
			Buckets: []float64{0.01, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"service", "code", "method"})
	httpResponseSize = prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "A histogram of HTTP response sizes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		}, []string{"service"})

	// exported metrics
	UserGauge = prom.NewGauge(
		prom.GaugeOpts{
			Name: "db_users",
			Help: "Gauge of users in the database",
		})
	GroupGauge = prom.NewGauge(
		prom.GaugeOpts{
			Name: "db_groups",
			Help: "Gauge of groups in the database",
		})
	MembershipGauge = prom.NewGauge(
		prom.GaugeOpts{
			Name: "db_memberships",
			Help: "Gauge of memberships in the database",
		})
	DatabaseQueryCounter = prom.NewCounter(
		prom.CounterOpts{
			Name: "db_queries_total",
			Help: "Counter of team.getConversationHistoryWithin calls",
		})
	DatabaseQueryLatencyHistogram = prom.NewHistogram(
		prom.HistogramOpts{
			Name:    "db_queries_latency_seconds",
			Help:    "A histogram of database query latencies in seconds",
			Buckets: []float64{0.01, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		})
)

func init() {
	prom.MustRegister(
		prom.NewBuildInfoCollector(), // tracks go build info

		// built in http request collectors
		httpRequestInFlightGauge,
		httpRequestCounter,
		httpRequestDuration,
		httpResponseSize,

		// exported collectors to be used throughout application
		UserGauge,
		GroupGauge,
		MembershipGauge,
		DatabaseQueryCounter,
		DatabaseQueryLatencyHistogram,
	)
}
