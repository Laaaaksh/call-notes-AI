package apperror

// ErrorResponse is the standardized API error response format.
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Code      string                 `json:"code,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}
