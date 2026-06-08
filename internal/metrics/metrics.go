package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TransactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_transactions_total",
			Help: "Total number of transaction attempts.",
		},
		[]string{"status", "error_type"},
	)

	TransactionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ledger_transaction_duration_seconds",
			Help:    "Duration of CreateTransaction calls in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)

	DBErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ledger_db_errors_total",
			Help: "Total number of unexpected database errors.",
		},
	)

	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ledger_grpc_active_connections",
			Help: "Number of active gRPC connections.",
		},
	)
)
