package constants

// Metric names
const (
	MetricRequestDuration   = "http_request_duration_seconds"
	MetricRequestTotal      = "http_requests_total"
	MetricActiveSessionsCount = "active_sessions_count"
	MetricFieldsExtracted   = "fields_extracted_total"
	MetricFieldsOverridden  = "fields_overridden_total"
	MetricSTTLatency        = "stt_latency_seconds"
	MetricExtractionLatency = "extraction_latency_seconds"
	MetricEndToEndLatency   = "end_to_end_latency_seconds"
	MetricLLMInvocations    = "llm_invocations_total"
	MetricDeepgramErrors    = "deepgram_errors_total"
	MetricSalesforceUpserts = "salesforce_upserts_total"
	MetricConfidenceAvg     = "field_confidence_average"
	MetricDBConnectionsOpen = "db_connections_open"
	MetricDBConnectionsIdle = "db_connections_idle"
)

// Metric labels
const (
	LabelMethod     = "method"
	LabelPath       = "path"
	LabelStatusCode = "status_code"
	LabelFieldName  = "field_name"
	LabelSource     = "source"
	LabelReason     = "reason"
)
