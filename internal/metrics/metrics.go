package metrics

import (
	"github.com/call-notes-ai-service/internal/constants"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    constants.MetricRequestDuration,
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{constants.LabelMethod, constants.LabelPath, constants.LabelStatusCode},
	)

	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: constants.MetricRequestTotal,
			Help: "Total number of HTTP requests",
		},
		[]string{constants.LabelMethod, constants.LabelPath, constants.LabelStatusCode},
	)

	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: constants.MetricActiveSessionsCount,
			Help: "Number of active call sessions",
		},
	)

	FieldsExtracted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: constants.MetricFieldsExtracted,
			Help: "Total fields extracted by source",
		},
		[]string{constants.LabelFieldName, constants.LabelSource},
	)

	FieldsOverridden = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: constants.MetricFieldsOverridden,
			Help: "Total fields overridden by agent",
		},
	)

	STTLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    constants.MetricSTTLatency,
			Help:    "Speech-to-text latency in seconds",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
	)

	ExtractionLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    constants.MetricExtractionLatency,
			Help:    "Entity extraction pipeline latency in seconds",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
	)

	EndToEndLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    constants.MetricEndToEndLatency,
			Help:    "End-to-end pipeline latency (speech to UI update)",
			Buckets: []float64{0.5, 1, 2, 3, 5, 10},
		},
	)

	LLMInvocations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: constants.MetricLLMInvocations,
			Help: "Total LLM invocations by reason",
		},
		[]string{constants.LabelReason},
	)

	DeepgramErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: constants.MetricDeepgramErrors,
			Help: "Total Deepgram STT errors",
		},
	)

	SalesforceUpserts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: constants.MetricSalesforceUpserts,
			Help: "Total Salesforce upsert attempts",
		},
		[]string{constants.LabelStatusCode},
	)

	DBConnectionsOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: constants.MetricDBConnectionsOpen,
			Help: "Number of open database connections",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: constants.MetricDBConnectionsIdle,
			Help: "Number of idle database connections",
		},
	)
)

func RecordHTTPRequest(method, path string, statusCode int, durationSeconds float64) {
	label := statusCodeToLabel(statusCode)
	HTTPRequestDuration.WithLabelValues(method, path, label).Observe(durationSeconds)
	HTTPRequestsTotal.WithLabelValues(method, path, label).Inc()
}

func RecordFieldExtracted(fieldName, source string) {
	FieldsExtracted.WithLabelValues(fieldName, source).Inc()
}

func RecordLLMInvocation(reason string) {
	LLMInvocations.WithLabelValues(reason).Inc()
}

func statusCodeToLabel(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}
