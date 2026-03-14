package constants

// SessionStatus is a typed string for session lifecycle states
type SessionStatus string

const (
	SessionStatusCreated     SessionStatus = "CREATED"
	SessionStatusActive      SessionStatus = "ACTIVE"
	SessionStatusInterrupted SessionStatus = "INTERRUPTED"
	SessionStatusReviewing   SessionStatus = "REVIEWING"
	SessionStatusSubmitted   SessionStatus = "SUBMITTED"
	SessionStatusCompleted   SessionStatus = "COMPLETED"
	SessionStatusErrored     SessionStatus = "ERRORED"
)

// String returns the string representation
func (s SessionStatus) String() string { return string(s) }

// ExtractionSource is a typed string for field extraction sources
type ExtractionSource string

const (
	SourceRuleEngine    ExtractionSource = "rule_engine"
	SourceMedicalNER    ExtractionSource = "medical_ner"
	SourceLLM           ExtractionSource = "llm"
	SourceAgentOverride ExtractionSource = "agent_override"
	SourceHistory       ExtractionSource = "history"
)

// String returns the string representation
func (e ExtractionSource) String() string { return string(e) }

// SpeakerLabel is a typed string for speaker identification
type SpeakerLabel string

const (
	SpeakerPatient SpeakerLabel = "patient"
	SpeakerAgent   SpeakerLabel = "agent"
	SpeakerUnknown SpeakerLabel = "unknown"
)

// String returns the string representation
func (s SpeakerLabel) String() string { return string(s) }

// Confidence thresholds
const (
	ConfidenceHigh   = 0.90
	ConfidenceMedium = 0.70
	ConfidenceLow    = 0.50
)

// Redis key prefixes for session state
const (
	RedisKeySessionState      = "session:%s:state"
	RedisKeySessionFields     = "session:%s:fields"
	RedisKeySessionTranscript = "session:%s:transcript"
)
