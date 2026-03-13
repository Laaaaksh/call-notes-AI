package health

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/call-notes-ai-service/internal/constants"
)

type HTTPHandler struct {
	core ICore
}

func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(constants.RouteHealthLive, h.LivenessCheck)
	r.Get(constants.RouteHealthReady, h.ReadinessCheck)
}

func (h *HTTPHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	status, code := h.core.RunLivenessCheck(r.Context())
	h.writeResponse(w, code, status)
}

func (h *HTTPHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	status, code := h.core.RunReadinessCheck(r.Context())
	h.writeResponse(w, code, status)
}

func (h *HTTPHandler) writeResponse(w http.ResponseWriter, statusCode int, status string) {
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: status})
}
