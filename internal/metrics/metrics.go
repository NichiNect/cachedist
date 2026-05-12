package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var NodeID = "unknown"

// Metrics definitions
var (
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of successful cache lookups",
		},
		[]string{"node_id"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of missed cache lookups",
		},
		[]string{"node_id"},
	)

	CacheEvictionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_evictions_total",
			Help: "Total number of cache evictions",
		},
		[]string{"node_id"},
	)

	CacheItemsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_items_total",
			Help: "Total number of items in the cache",
		},
		[]string{"node_id"},
	)

	CacheMemoryBytes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_memory_bytes",
			Help: "Estimated memory usage of cache items in bytes",
		},
		[]string{"node_id"},
	)

	CacheRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_request_duration_seconds",
			Help:    "Histogram of cache request durations",
			Buckets: []float64{0.0001, 0.001, 0.01, 0.1},
		},
		[]string{"node_id", "operation"},
	)

	ReplicationSuccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "replication_success_total",
			Help: "Total number of successful key replications",
		},
		[]string{"node_id"},
	)

	ReplicationFailureTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "replication_failure_total",
			Help: "Total number of failed key replications",
		},
		[]string{"node_id"},
	)
)

// Helper functions that automatically apply the NodeID label
func IncHit() {
	CacheHitsTotal.WithLabelValues(NodeID).Inc()
}

func IncMiss() {
	CacheMissesTotal.WithLabelValues(NodeID).Inc()
}

func IncEviction() {
	CacheEvictionsTotal.WithLabelValues(NodeID).Inc()
}

func AddItems(count float64) {
	CacheItemsTotal.WithLabelValues(NodeID).Add(count)
}

func AddMemory(bytes float64) {
	CacheMemoryBytes.WithLabelValues(NodeID).Add(bytes)
}

func ObserveRequest(operation string, durationSeconds float64) {
	CacheRequestDurationSeconds.WithLabelValues(NodeID, operation).Observe(durationSeconds)
}

func IncReplicationSuccess() {
	ReplicationSuccessTotal.WithLabelValues(NodeID).Inc()
}

func IncReplicationFailure() {
	ReplicationFailureTotal.WithLabelValues(NodeID).Inc()
}
