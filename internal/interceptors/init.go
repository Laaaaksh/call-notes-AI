package interceptors

import (
	"net/http"
	"time"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/go-chi/chi/v5/middleware"
)

var DefaultRequestTimeout = time.Duration(constants.DefaultRequestTimeoutSeconds) * time.Second

func DefaultMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Timeout(DefaultRequestTimeout),
	}
}
