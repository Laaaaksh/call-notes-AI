package apperror

// Public error messages — user-facing
const (
	MsgInternalError     = "An internal error occurred. Please try again later."
	MsgInvalidRequest    = "The request is invalid."
	MsgResourceNotFound  = "The specified resource was not found."
	MsgDuplicateResource = "A resource with this identifier already exists."
	MsgServiceUnavailable = "Service is temporarily unavailable."
	MsgInvalidJSONBody   = "Invalid JSON in request body."
	MsgRequestTimeout    = "Request timed out."
	MsgRateLimited       = "Too many requests. Please try again later."
	MsgUnauthorized      = "Unauthorized access."

	MsgSessionNotFound     = "The specified call session was not found."
	MsgSessionAlreadyExists = "A session with this call ID already exists."
	MsgInvalidSessionID    = "Session ID must be a valid UUID."
	MsgInvalidFieldName    = "Invalid field name provided."
	MsgTriageNotFound      = "Triage assessment not found for this session."
	MsgFollowUpNotFound    = "Follow-up not found."
	MsgPatientNotFound     = "No patient history found for this phone number."
)

// Error field keys for contextual information
const (
	FieldSessionID    = "session_id"
	FieldAgentID      = "agent_id"
	FieldCallID       = "call_id"
	FieldFieldName    = "field_name"
	FieldPatientPhone = "patient_phone"
	FieldFollowUpID   = "followup_id"
	FieldRequestID    = "request_id"
)
