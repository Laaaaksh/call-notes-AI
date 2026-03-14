// Package interceptors provides HTTP middleware chain configuration.
package interceptors

import (
	"net/http"
	"time"

	"github.com/call-notes-ai-service/internal/config"
	"github.com/call-notes-ai-service/internal/constants"
	"github.com/go-chi/chi/v5/middleware"
)

// DefaultRequestTimeout is the default timeout for HTTP requests
var DefaultRequestTimeout = time.Duration(constants.DefaultRequestTimeoutSeconds) * time.Second

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain represents an ordered list of middleware
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{middlewares: middlewares}
}

// Then wraps the final handler with all middleware in the chain
func (c *Chain) Then(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

// Append adds middleware to the chain
func (c *Chain) Append(middlewares ...Middleware) *Chain {
	newMiddlewares := make([]Middleware, 0, len(c.middlewares)+len(middlewares))
	newMiddlewares = append(newMiddlewares, c.middlewares...)
	newMiddlewares = append(newMiddlewares, middlewares...)
	return &Chain{middlewares: newMiddlewares}
}

// DefaultMiddleware returns the standard middleware chain for the main API
func DefaultMiddleware() []Middleware {
	return []Middleware{
		middleware.RequestID,
		middleware.RealIP,
		RecoveryMiddleware,
		RequestIDMiddleware,
		MetricsMiddleware,
		RequestLoggerMiddleware,
	}
}

// DefaultMiddlewareWithTimeout returns the standard middleware chain with timeout
func DefaultMiddlewareWithTimeout(timeout time.Duration) []Middleware {
	return []Middleware{
		middleware.RequestID,
		middleware.RealIP,
		RecoveryMiddleware,
		RequestIDMiddleware,
		TimeoutMiddleware(timeout),
		MetricsMiddleware,
		RequestLoggerMiddleware,
	}
}

// ApplyMiddleware applies a list of middleware to a handler
func ApplyMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	chain := NewChain(middlewares...)
	return chain.Then(handler)
}

// GetChiMiddleware returns middleware compatible with chi.Router.Use()
func GetChiMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		RecoveryMiddleware,
		RequestIDMiddleware,
		SecurityHeadersMiddleware,
		MaxBytesMiddleware,
		ContentTypeValidationMiddleware,
		TimeoutMiddleware(DefaultRequestTimeout),
		MetricsMiddleware,
		RequestLoggerMiddleware,
	}
}

// MiddlewareConfig holds all middleware configuration
type MiddlewareConfig struct {
	RateLimit config.RateLimitConfig
	Tracing   config.TracingConfig
}

// GetChiMiddlewareWithFullConfig returns the production-ready middleware chain.
// Middleware order is carefully designed:
// 1. RequestID — generate/propagate request ID first for tracing
// 2. RealIP — extract real client IP from proxy headers
// 3. TracingMiddleware — OpenTelemetry distributed tracing (early for full span coverage)
// 4. TraceContextMiddleware — add trace IDs to context for logging
// 5. RecoveryMiddleware — panic recovery (must be early to catch all panics)
// 6. RequestIDMiddleware — add request ID to response headers
// 7. SecurityHeadersMiddleware — add security headers early
// 8. RateLimitMiddleware — rate limiting before heavy processing
// 9. MaxBytesMiddleware — limit body size before parsing (DoS protection)
// 10. ContentTypeValidationMiddleware — validate content-type before parsing
// 11. TimeoutMiddleware — request timeout protection
// 12. MetricsMiddleware — record metrics (after timeout to measure actual time)
// 13. RequestLoggerMiddleware — log requests (captures response details)
func GetChiMiddlewareWithFullConfig(cfg MiddlewareConfig) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		TracingMiddleware(cfg.Tracing),
		TraceContextMiddleware,
		RecoveryMiddleware,
		RequestIDMiddleware,
		SecurityHeadersMiddleware,
		RateLimitMiddleware(cfg.RateLimit),
		MaxBytesMiddleware,
		ContentTypeValidationMiddleware,
		TimeoutMiddleware(DefaultRequestTimeout),
		MetricsMiddleware,
		RequestLoggerMiddleware,
	}
}
