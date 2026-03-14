package interceptors

import (
	"context"
	"net/http"

	"github.com/call-notes-ai-service/internal/config"
	"github.com/call-notes-ai-service/internal/constants/contextkeys"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracingSpanHTTPRequest   = "http-request"
	tracingSpanNameSeparator = " "
	tracingAttrRequestID     = "request_id"
	tracingAttrUserAgent     = "http.user_agent"
	tracingAttrClientIP      = "http.client_ip"
)

// TracingMiddleware creates an OpenTelemetry HTTP tracing middleware
func TracingMiddleware(cfg config.TracingConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return createPassThroughMiddleware()
	}
	return createTracingHandler()
}

func createTracingHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := otelhttp.NewHandler(next, tracingSpanHTTPRequest,
			otelhttp.WithSpanNameFormatter(formatSpanName),
		)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
		})
	}
}

// TraceContextMiddleware extracts trace/span IDs and adds them to context for logging
func TraceContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = addTraceIDsToContext(ctx)
		addCustomSpanAttributes(r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func addTraceIDsToContext(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ctx
	}

	if span.SpanContext().HasTraceID() {
		ctx = context.WithValue(ctx, contextkeys.TraceID, span.SpanContext().TraceID().String())
	}
	if span.SpanContext().HasSpanID() {
		ctx = context.WithValue(ctx, contextkeys.SpanID, span.SpanContext().SpanID().String())
	}
	return ctx
}

func formatSpanName(_ string, r *http.Request) string {
	return r.Method + tracingSpanNameSeparator + r.URL.Path
}

func addCustomSpanAttributes(r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	if !span.IsRecording() {
		return
	}

	if requestID, ok := r.Context().Value(contextkeys.RequestID).(string); ok {
		span.SetAttributes(attribute.String(tracingAttrRequestID, requestID))
	}

	if userAgent := r.UserAgent(); userAgent != "" {
		span.SetAttributes(attribute.String(tracingAttrUserAgent, userAgent))
	}

	if clientIP := r.RemoteAddr; clientIP != "" {
		span.SetAttributes(attribute.String(tracingAttrClientIP, clientIP))
	}
}
