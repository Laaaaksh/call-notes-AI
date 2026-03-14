package constants

// HTTP headers and content types
const (
	HeaderContentType = "Content-Type"
	ContentTypeJSON   = "application/json"
)

// HTTP status codes
const (
	HTTPStatusOK                 = 200
	HTTPStatusCreated            = 201
	HTTPStatusBadRequest         = 400
	HTTPStatusUnauthorized       = 401
	HTTPStatusNotFound           = 404
	HTTPStatusConflict           = 409
	HTTPStatusTooManyRequests    = 429
	HTTPStatusServiceUnavailable = 503
)
