package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "otlp_log_parser_assignment_requests_total",
		Help: "Total number of OTLP log export requests received.",
	})

	LogRecordsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "otlp_log_parser_assignment_log_records_processed_total",
		Help: "Total number of log records processed.",
	})

	AttributeValuesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "otlp_log_parser_assignment_attribute_values_total",
		Help: "Total number of times each attribute value has been seen.",
	}, []string{"value"})
)
