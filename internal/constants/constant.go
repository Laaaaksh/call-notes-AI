package constants

// Application environment
const (
	EnvKeyAppEnv  = "APP_ENV"
	EnvDefaultDev = "dev"
)

// Server names
const (
	ServerNameMain = "main"
	ServerNameOps  = "ops"
)

// API routing
const (
	APIVersionPrefix = "/v1"
)

// Request timeout
const (
	DefaultRequestTimeoutSeconds = 30
)

// Health status
const (
	StatusServing    = "SERVING"
	StatusNotServing = "NOT_SERVING"
)

// Response field values
const (
	ResponseStatusUpdated = "updated"
	ResponseStatusPurged  = "purged"
	ResponseKeyError      = "error"
	ResponseKeyStatus     = "status"
)

// LLM reasoning trigger values
const (
	TriggerAmbiguityOrCorrection = "ambiguity_or_correction"
)
