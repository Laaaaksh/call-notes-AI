package triage

import (
	"net/http"

	"github.com/call-notes-ai-service/internal/modules/triage/entities"
	"github.com/call-notes-ai-service/internal/utils"
	"github.com/call-notes-ai-service/pkg/apperror"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// HTTPHandler handles triage HTTP requests
type HTTPHandler struct {
	core ICore
}

// NewHTTPHandler creates a new triage HTTP handler
func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

// RegisterRoutes registers triage routes on the router
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(entities.RouteSessionTriage, h.GetTriage)
}

// GetTriage handles GET /sessions/{sessionID}/triage
func (h *HTTPHandler) GetTriage(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		utils.WriteError(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, err, apperror.MsgInvalidSessionID))
		return
	}

	assessment, err := h.core.GetAssessment(r.Context(), sessionID)
	if err != nil {
		utils.WriteError(w, r, apperror.NewWithMessage(apperror.CodeNotFound, err, apperror.MsgTriageNotFound))
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

	utils.WriteJSON(w, http.StatusOK, resp)
}
