package triage

import (
	"encoding/json"
	"net/http"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/modules/triage/entities"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type HTTPHandler struct {
	core ICore
}

func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(entities.RouteSessionTriage, h.GetTriage)
}

func (h *HTTPHandler) GetTriage(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrMsgInvalidSessionID)
		return
	}

	assessment, err := h.core.GetAssessment(r.Context(), sessionID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "triage assessment not found")
		return
	}

	resp := &entities.TriageResponse{
		SessionID:      sessionID.String(),
		UrgencyLevel:   assessment.UrgencyLevel,
		CompositeScore: assessment.CompositeScore,
		Symptoms:       assessment.Symptoms,
		RedFlags:       assessment.RedFlags,
		Modifiers:      assessment.ModifiersApplied,
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func (h *HTTPHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
