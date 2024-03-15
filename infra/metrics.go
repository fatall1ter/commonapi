package infra

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	common = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_up",
			Help: "Common metric of state of available service and subservices",
		},
		[]string{"scope", "destination", "version", "githash", "build"},
	)

	httpDuration = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "api_request_duration_nanosecond",
			Help:       "Summary of api req-resp durations in nanoseconds by quantile",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"url", "code"},
	)

	redis = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_key_items_total",
			Help: "Redis key's count items",
		},
		[]string{"key"},
	)
)
