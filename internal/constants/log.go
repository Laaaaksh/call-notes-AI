package constants

// Log format and levels
const (
	LogFormatJSON        = "json"
	LogLevelDebug        = "debug"
	LogLevelInfo         = "info"
	LogLevelWarn         = "warn"
	LogLevelError        = "error"
	LogOutputStdout      = "stdout"
	LogOutputStderr      = "stderr"
	LogEncoderTimeKey    = "timestamp"
	LogEncoderMessageKey = "msg"
	LogEncoderLevelKey   = "level"
	LogEncoderCallerKey  = "caller"
)

// Log field keys — core
const (
	LogKeyError          = "error"
	LogKeyRequestID      = "request_id"
	LogFieldName         = "name"
	LogFieldEnv          = "env"
	LogFieldPort         = "port"
	LogFieldOpsPort      = "ops_port"
	LogFieldAddr         = "addr"
	LogFieldTraceID      = "trace_id"
	LogFieldSpanID       = "span_id"
	LogFieldSessionID    = "session_id"
	LogFieldAgentID      = "agent_id"
	LogFieldCallID       = "call_id"
	LogFieldFieldName    = "field_name"
	LogFieldConfidence   = "confidence"
	LogFieldSource       = "source"
	LogFieldDelaySeconds = "delay_seconds"
	LogFieldClientID     = "client_id"
	LogFieldTrigger      = "trigger"
	LogFieldAttempt      = "attempt"
	LogFieldMaxRetries   = "max_retries"
	LogFieldBackoff      = "backoff"
	LogFieldHost         = "host"
	LogFieldDatabase     = "database"
	LogFieldMaxConns     = "max_connections"
	LogFieldModel        = "model"
	LogFieldLatencyMs    = "latency_ms"
	LogFieldInstanceURL  = "instance_url"
	LogFieldObject       = "object"
	LogFieldExternalID   = "external_id"
	LogFieldRecordID     = "record_id"
	LogFieldSequence     = "sequence"
	LogFieldIsFinal      = "is_final"
	LogFieldSpeaker      = "speaker"
)

// Log field keys — futuristic modules
const (
	LogFieldPatientPhone   = "patient_phone"
	LogFieldPredictedCount = "predicted_count"
	LogFieldSessionCount   = "session_count"
	LogFieldEmotionType    = "emotion_type"
	LogFieldIntensity      = "intensity"
	LogFieldUrgencyLevel   = "urgency_level"
	LogFieldTriageScore    = "triage_score"
	LogFieldFollowupType   = "followup_type"
	LogFieldFollowupID     = "followup_id"
	LogFieldFollowupStatus = "followup_status"
	LogFieldDueDate        = "due_date"
)

// Log messages — boot and server lifecycle
const (
	LogMsgStartingService          = "Starting service"
	LogMsgMainServerStarting       = "Main API server starting"
	LogMsgOpsServerStarting        = "Ops server starting"
	LogMsgMainServerFailed         = "Main server failed"
	LogMsgOpsServerFailed          = "Ops server failed"
	LogMsgShutdownSignalReceived   = "Shutdown signal received"
	LogMsgServiceStopped           = "Service stopped"
	LogMsgServiceMarkedUnhealthy   = "Service marked as unhealthy"
	LogMsgWaitingForShutdownDelay  = "Waiting for connection drain"
	LogMsgGracefulShutdownComplete = "Graceful shutdown complete"
	LogMsgShutdownTimeoutExceeded  = "Shutdown timeout exceeded"
	LogMsgMainServerShutdownErr    = "Main server shutdown error"
	LogMsgOpsServerShutdownErr     = "Ops server shutdown error"
	LogMsgMainServerShutdownDone   = "Main server shutdown complete"
	LogMsgOpsServerShutdownDone    = "Ops server shutdown complete"
	LogMsgTracerInitFailed         = "Tracer initialization failed"
	LogMsgFailedToInitDB           = "Failed to initialize database"
	LogMsgReadinessCheckFailed     = "Readiness check failed"
	LogMsgFailedToEncodeResponse   = "Failed to encode response"
)

// Log messages — session
const (
	LogMsgSessionCreated      = "Call session created"
	LogMsgSessionCreateFailed = "Failed to create session"
	LogMsgSessionGetFailed    = "Failed to get session"
	LogMsgSessionEnded        = "Call session ended"
	LogMsgSessionPurgeFailed  = "Session purge failed"
	LogMsgSessionPurged       = "Session purged (DPDP right to erasure)"
	LogMsgFieldExtracted      = "Field extracted"
	LogMsgFieldUpdated        = "Field updated"
	LogMsgAgentOverride       = "Agent override applied"
	LogMsgSalesforceSubmitted = "Record submitted to Salesforce"
)

// Log messages — infrastructure
const (
	LogMsgRedisConnFailed   = "Redis connection failed, session caching disabled"
	LogMsgRedisConnected    = "Redis connected"
	LogMsgDBPoolInitialized = "Database pool initialized"
	LogMsgDBConnAttempt     = "Database connection attempt"
	LogMsgDBConnected       = "Database connected"
	LogMsgDBConnFailed      = "Database connection failed, retrying"
	LogMsgDBPoolClosed      = "Database pool closed"
)

// Log messages — external services
const (
	LogMsgDeepgramConnected    = "Deepgram WebSocket connected"
	LogMsgDeepgramDisconnected = "Deepgram WebSocket disconnected"
	LogMsgDeepgramReadErr      = "Deepgram read error"
	LogMsgLLMInvoked           = "LLM reasoning invoked"
	LogMsgLLMComplete          = "LLM invocation complete"
	LogMsgLLMReasonFailed      = "LLM reasoning failed, continuing with L1+L2"
	LogMsgLLMConflictFailed    = "LLM conflict resolution failed"
	LogMsgLLMSummaryFailed     = "LLM summary generation failed"
	LogMsgSFAuthenticated      = "Salesforce authenticated"
	LogMsgSFUpsertComplete     = "Salesforce upsert complete"
	LogMsgSFUpsertFailed       = "Salesforce upsert failed"
	LogMsgSFCaseUpserted       = "Salesforce case upserted"
	LogMsgPIIRedacted          = "PII redacted from transcript"
)

// Log messages — WebSocket
const (
	LogMsgWSClientConnected    = "WebSocket client connected"
	LogMsgWSClientDisconnected = "WebSocket client disconnected"
)

// Log messages — transcription
const (
	LogMsgTranscriptChunkProcessed = "Transcript chunk processed"
)

// Log messages — prediction module
const (
	LogMsgPredictionLookupFailed = "Patient history lookup failed"
	LogMsgPredictionComplete     = "Predictive pre-population complete"
)

// Log messages — sentiment module
const (
	LogMsgSentimentLogFailed      = "Failed to log sentiment"
	LogMsgSentimentAlertTriggered = "Sentiment alert triggered"
)

// Log messages — triage module
const (
	LogMsgTriageAssessmentFailed = "Triage assessment persistence failed"
	LogMsgTriageUpdated          = "Triage assessment updated"
)

// Log messages — follow-up module
const (
	LogMsgFollowupDetected     = "Follow-up detected in transcript"
	LogMsgFollowupCreateFailed = "Failed to create follow-up"
	LogMsgFollowupConfirmed    = "Follow-up confirmed by agent"
)

// Log messages — analytics module
const (
	LogMsgAnalyticsQueryFailed = "Analytics query failed"
)
