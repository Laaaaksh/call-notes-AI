package interceptors

import (
	"fmt"
	"net/http"

	"github.com/call-notes-ai-service/internal/config"
	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"golang.org/x/time/rate"
)

const (
	defaultRetryAfterSeconds = "1"
)

// RateLimitMiddleware creates a rate limiting middleware using token bucket algorithm.
// If rate limiting is disabled in config, it returns a pass-through middleware.
func RateLimitMiddleware(cfg config.RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return createPassThroughMiddleware()
	}

	limiter := createLimiter(cfg)
	return createRateLimitHandler(limiter)
}

func createPassThroughMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}

func createLimiter(cfg config.RateLimitConfig) *rate.Limiter {
	return rate.NewLimiter(rate.Limit(cfg.RequestsPerSec), cfg.BurstSize)
}

func createRateLimitHandler(limiter *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				handleRateLimitExceeded(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func handleRateLimitExceeded(w http.ResponseWriter, r *http.Request) {
	logRateLimitExceeded(r)
	writeRateLimitResponse(w)
}

func logRateLimitExceeded(r *http.Request) {
	logger.Ctx(r.Context()).Warnw("Rate limit exceeded",
		constants.LogKeyError, "too many requests",
		"method", r.Method,
		"path", r.URL.Path,
	)
}

func writeRateLimitResponse(w http.ResponseWriter) {
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.Header().Set(headerRetryAfter, defaultRetryAfterSeconds)
	w.WriteHeader(http.StatusTooManyRequests)
	body := fmt.Sprintf(`{"error":"%s","code":"%s"}`, errMsgRateLimitExceeded, errCodeRateLimited)
	_, _ = w.Write([]byte(body))
}
