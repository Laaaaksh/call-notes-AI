package health

import (
	"net/http"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/utils"
	"github.com/go-chi/chi/v5"
)

// HTTPHandler handles health check HTTP requests
type HTTPHandler struct {
	core ICore
}

// HealthResponse is the health check response body
type HealthResponse struct {
	Status string `json:"status"`
}

// NewHTTPHandler creates a new health HTTP handler
func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

// RegisterRoutes registers health routes on the router
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(constants.RouteHealthLive, h.LivenessCheck)
	r.Get(constants.RouteHealthReady, h.ReadinessCheck)
}

// LivenessCheck handles GET /health/live
func (h *HTTPHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	status, code := h.core.RunLivenessCheck(r.Context())
	utils.WriteJSON(w, code, HealthResponse{Status: status})
}

// ReadinessCheck handles GET /health/ready
func (h *HTTPHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	status, code := h.core.RunReadinessCheck(r.Context())
	utils.WriteJSON(w, code, HealthResponse{Status: status})
}
