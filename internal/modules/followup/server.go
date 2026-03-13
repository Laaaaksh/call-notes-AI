package followup

import (
	"encoding/json"
	"net/http"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/modules/followup/entities"
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
	r.Get(entities.RouteSessionFollowups, h.GetFollowUps)
	r.Post(entities.RouteSessionFollowupsConfirm, h.ConfirmFollowUp)
}

func (h *HTTPHandler) GetFollowUps(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrMsgInvalidSessionID)
		return
	}

	followups, err := h.core.GetFollowUps(r.Context(), sessionID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if followups == nil {
		followups = []entities.FollowUp{}
	}

	h.writeJSON(w, http.StatusOK, &entities.FollowUpResponse{FollowUps: followups})
}

func (h *HTTPHandler) ConfirmFollowUp(w http.ResponseWriter, r *http.Request) {
	var req entities.ConfirmFollowUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrMsgInvalidRequestBody)
		return
	}

	fu, err := h.core.ConfirmFollowUp(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, fu)
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
