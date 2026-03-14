package interceptors

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/metrics"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	stackSizeBytes    = 4096
	maxRequestBodyMB  = 1 << 20 // 1 MB
	uuidLength        = 36
	pathPlaceholderID   = ":id"
	pathPlaceholderUUID = ":uuid"

	headerRequestID          = "X-Request-ID"
	headerXContentTypeOpts   = "X-Content-Type-Options"
	headerXFrameOptions      = "X-Frame-Options"
	headerCacheControl       = "Cache-Control"
	headerRetryAfter         = "Retry-After"
	headerAccessControlOrigin  = "Access-Control-Allow-Origin"
	headerAccessControlMethods = "Access-Control-Allow-Methods"
	headerAccessControlHeaders = "Access-Control-Allow-Headers"

	valueNoSniff = "nosniff"
	valueDeny    = "DENY"
	valueNoStore = "no-store"

	corsAllowOriginAll     = "*"
	corsAllowMethodsAll    = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowHeadersCommon = "Content-Type, Authorization, X-Request-ID"

	contentTypeJSONPrefix = "application/json"

	httpStatusUnsupportedMediaType = 415

	errMsgInternalServerError = "Internal server error"
	errMsgRequestTimeout      = "Request timeout"
	errMsgUnsupportedMediaType = "Content-Type must be application/json"
	errMsgRateLimitExceeded    = "Rate limit exceeded"

	errCodeInternalError    = "INTERNAL_ERROR"
	errCodeTimeout          = "TIMEOUT"
	errCodeUnsupportedMedia = "UNSUPPORTED_MEDIA_TYPE"
	errCodeRateLimited      = "RATE_LIMITED"

	logMsgPanicRecovered    = "Panic recovered"
	logMsgHTTPRequest       = "HTTP request completed"
	logMsgInvalidContentType = "Invalid content type"

	logKeyStack        = "stack"
	logKeyPath         = "path"
	logKeyMethod       = "method"
	logKeyStatusCode   = "status_code"
	logKeyDuration     = "duration_ms"
	logKeyBytesWritten = "bytes_written"
)

// RecoveryMiddleware recovers from panics and logs the stack trace
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(r, w)
		next.ServeHTTP(w, r)
	})
}

func recoverFromPanic(r *http.Request, w http.ResponseWriter) {
	err := recover()
	if err == nil {
		return
	}

	stack := captureStackTrace()
	logPanicRecovery(r, err, stack)
	sendInternalServerError(w, r)
}

func captureStackTrace() string {
	stack := make([]byte, stackSizeBytes)
	length := runtime.Stack(stack, false)
	return string(stack[:length])
}

func logPanicRecovery(r *http.Request, err interface{}, stack string) {
	logger.Ctx(r.Context()).Errorw(logMsgPanicRecovered,
		constants.LogKeyError, err,
		logKeyStack, stack,
		logKeyPath, r.URL.Path,
		logKeyMethod, r.Method,
	)
}

func sendInternalServerError(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(http.StatusInternalServerError)

	response := buildErrorResponseJSON(errMsgInternalServerError, errCodeInternalError, requestID)
	_, _ = w.Write([]byte(response))
}

func buildErrorResponseJSON(errMsg, errCode, requestID string) string {
	if requestID != "" {
		return fmt.Sprintf(`{"error":"%s","code":"%s","request_id":"%s"}`, errMsg, errCode, requestID)
	}
	return fmt.Sprintf(`{"error":"%s","code":"%s"}`, errMsg, errCode)
}

// RequestLoggerMiddleware logs HTTP requests with timing information
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		logHTTPRequest(r, ww, time.Since(start))
	})
}

func logHTTPRequest(r *http.Request, ww middleware.WrapResponseWriter, duration time.Duration) {
	logger.Ctx(r.Context()).Infow(logMsgHTTPRequest,
		logKeyMethod, r.Method,
		logKeyPath, r.URL.Path,
		logKeyStatusCode, ww.Status(),
		logKeyDuration, duration.Milliseconds(),
		logKeyBytesWritten, ww.BytesWritten(),
	)
}

// MetricsMiddleware records HTTP request metrics to Prometheus
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		recordRequestMetrics(r, ww.Status(), time.Since(start))
	})
}

func recordRequestMetrics(r *http.Request, statusCode int, duration time.Duration) {
	path := normalizePath(r.URL.Path)
	metrics.RecordHTTPRequest(r.Method, path, statusCode, duration.Seconds())
}

// normalizePath replaces dynamic segments (IDs, UUIDs) with placeholders
// to prevent high-cardinality metrics
func normalizePath(path string) string {
	segments := splitPath(path)
	for i, segment := range segments {
		segments[i] = normalizeSegment(segment)
	}
	return joinPath(segments)
}

func normalizeSegment(segment string) string {
	if isNumericSegment(segment) {
		return pathPlaceholderID
	}
	if isUUIDSegment(segment) {
		return pathPlaceholderUUID
	}
	return segment
}

func splitPath(path string) []string {
	if path == "" || path == "/" {
		return []string{}
	}
	trimmed := strings.TrimPrefix(path, "/")
	return strings.Split(trimmed, "/")
}

func joinPath(segments []string) string {
	if len(segments) == 0 {
		return "/"
	}
	return "/" + strings.Join(segments, "/")
}

func isNumericSegment(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isUUIDSegment(s string) bool {
	if len(s) != uuidLength {
		return false
	}
	for i, c := range s {
		if isUUIDDashPosition(i) {
			if c != '-' {
				return false
			}
			continue
		}
		if !isHexChar(c) {
			return false
		}
	}
	return true
}

func isUUIDDashPosition(position int) bool {
	return position == 8 || position == 13 || position == 18 || position == 23
}

func isHexChar(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// RequestIDMiddleware adds a request ID to the response header
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		if requestID != "" {
			w.Header().Set(headerRequestID, requestID)
		}
		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		body := fmt.Sprintf(`{"error":"%s","code":"%s"}`, errMsgRequestTimeout, errCodeTimeout)
		return http.TimeoutHandler(next, timeout, body)
	}
}

// CORSMiddleware adds basic CORS headers with wildcard origin.
// For production, use CORSMiddlewareWithOrigin with a specific origin.
func CORSMiddleware(next http.Handler) http.Handler {
	return CORSMiddlewareWithOrigin(corsAllowOriginAll)(next)
}

// CORSMiddlewareWithOrigin creates a CORS middleware with a configurable origin
func CORSMiddlewareWithOrigin(allowOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(headerAccessControlOrigin, allowOrigin)
			w.Header().Set(headerAccessControlMethods, corsAllowMethodsAll)
			w.Header().Set(headerAccessControlHeaders, corsAllowHeadersCommon)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeMiddleware ensures JSON content type for API responses
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds standard security headers to responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerXContentTypeOpts, valueNoSniff)
		w.Header().Set(headerXFrameOptions, valueDeny)
		w.Header().Set(headerCacheControl, valueNoStore)
		next.ServeHTTP(w, r)
	})
}

// MaxBytesMiddleware limits the size of request bodies to prevent DoS attacks
func MaxBytesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyMB)
		next.ServeHTTP(w, r)
	})
}

// ContentTypeValidationMiddleware validates Content-Type header for mutating requests
func ContentTypeValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isMutatingRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		if !isValidJSONContentType(r) {
			writeUnsupportedMediaTypeError(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isMutatingRequest(r *http.Request) bool {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

func isValidJSONContentType(r *http.Request) bool {
	contentType := r.Header.Get(constants.HeaderContentType)
	return strings.HasPrefix(contentType, contentTypeJSONPrefix)
}

func writeUnsupportedMediaTypeError(w http.ResponseWriter, r *http.Request) {
	logger.Ctx(r.Context()).Warnw(logMsgInvalidContentType,
		logKeyMethod, r.Method,
		logKeyPath, r.URL.Path,
		constants.HeaderContentType, r.Header.Get(constants.HeaderContentType),
	)

	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(httpStatusUnsupportedMediaType)
	body := fmt.Sprintf(`{"error":"%s","code":"%s"}`, errMsgUnsupportedMediaType, errCodeUnsupportedMedia)
	_, _ = w.Write([]byte(body))
}
