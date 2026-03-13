package contextkeys

type contextKey string

const (
	RequestID contextKey = "request_id"
	TraceID   contextKey = "trace_id"
	SpanID    contextKey = "span_id"
	SessionID contextKey = "session_id"
	AgentID   contextKey = "agent_id"
)
