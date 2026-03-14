package followup

import (
	"net/http"

	"github.com/call-notes-ai-service/internal/modules/followup/entities"
	"github.com/call-notes-ai-service/internal/utils"
	"github.com/call-notes-ai-service/pkg/apperror"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// HTTPHandler handles follow-up HTTP requests
type HTTPHandler struct {
	core ICore
}

// NewHTTPHandler creates a new follow-up HTTP handler
func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

// RegisterRoutes registers follow-up routes on the router
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(entities.RouteSessionFollowups, h.GetFollowUps)
	r.Post(entities.RouteSessionFollowupsConfirm, h.ConfirmFollowUp)
}

// GetFollowUps handles GET /sessions/{sessionID}/followups
func (h *HTTPHandler) GetFollowUps(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		utils.WriteError(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, err, apperror.MsgInvalidSessionID))
		return
	}

	followups, err := h.core.GetFollowUps(r.Context(), sessionID)
	if err != nil {
		utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
		return
	}

	if followups == nil {
		followups = []entities.FollowUp{}
	}

	utils.WriteJSON(w, http.StatusOK, &entities.FollowUpResponse{FollowUps: followups})
}

// ConfirmFollowUp handles POST /sessions/{sessionID}/followups/{followupID}/confirm
func (h *HTTPHandler) ConfirmFollowUp(w http.ResponseWriter, r *http.Request) {
	var req entities.ConfirmFollowUpRequest
	if decErr := utils.DecodeJSON(r, &req); decErr != nil {
		utils.WriteError(w, r, decErr)
		return
	}

	fu, err := h.core.ConfirmFollowUp(r.Context(), &req)
	if err != nil {
		utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, fu)
}
