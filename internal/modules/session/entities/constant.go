package entities

const (
	RouteSessionsCreate    = "/sessions"
	RouteSessionsGet       = "/sessions/{sessionID}"
	RouteSessionsFields    = "/sessions/{sessionID}/fields"
	RouteSessionsSubmit    = "/sessions/{sessionID}/submit"
	RouteSessionsTranscript = "/sessions/{sessionID}/transcript"
	RouteSessionsPurge     = "/sessions/{sessionID}/purge"

	ErrMsgSessionNotFound      = "session not found"
	ErrMsgSessionAlreadyActive = "agent already has an active session"
	ErrMsgInvalidSessionState  = "invalid session state transition"
	ErrMsgCallIDRequired       = "talkdesk call ID is required"
	ErrMsgAgentIDRequired      = "agent ID is required"
)
