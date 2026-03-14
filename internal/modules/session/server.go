package session

import (
	"net/http"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/modules/session/entities"
	"github.com/call-notes-ai-service/internal/utils"
	"github.com/call-notes-ai-service/pkg/apperror"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// HTTPHandler handles session HTTP requests
type HTTPHandler struct {
	core ICore
}

// StatusResponse is a generic status response
type StatusResponse struct {
	Status string `json:"status"`
}

// NewHTTPHandler creates a new session HTTP handler
func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

// RegisterRoutes registers session routes on the router
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Post(entities.RouteSessionsCreate, h.CreateSession)
	r.Get(entities.RouteSessionsGet, h.GetSession)
	r.Patch(entities.RouteSessionsFields, h.UpdateFields)
	r.Post(entities.RouteSessionsSubmit, h.SubmitSession)
	r.Delete(entities.RouteSessionsPurge, h.PurgeSession)
}

// CreateSession handles POST /sessions
func (h *HTTPHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req entities.StartSessionRequest
	if decErr := utils.DecodeJSON(r, &req); decErr != nil {
		utils.WriteError(w, r, decErr)
		return
	}

	resp, err := h.core.StartSession(r.Context(), &req)
	if err != nil {
		utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// GetSession handles GET /sessions/{sessionID}
func (h *HTTPHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID, appErr := parseSessionID(r)
	if appErr != nil {
		utils.WriteError(w, r, appErr)
		return
	}

	state, err := h.core.GetSessionState(r.Context(), sessionID)
	if err != nil {
		utils.WriteError(w, r, apperror.NewWithMessage(apperror.CodeNotFound, err, apperror.MsgSessionNotFound))
		return
	}

	utils.WriteJSON(w, http.StatusOK, state)
}

// UpdateFields handles PATCH /sessions/{sessionID}/fields
func (h *HTTPHandler) UpdateFields(w http.ResponseWriter, r *http.Request) {
	sessionID, appErr := parseSessionID(r)
	if appErr != nil {
		utils.WriteError(w, r, appErr)
		return
	}

	var req entities.UpdateFieldsRequest
	if decErr := utils.DecodeJSON(r, &req); decErr != nil {
		utils.WriteError(w, r, decErr)
		return
	}

	for _, o := range req.Overrides {
		if err := h.core.ApplyAgentOverride(r.Context(), sessionID, o.FieldName, o.Value); err != nil {
			utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, &StatusResponse{Status: constants.ResponseStatusUpdated})
}

// SubmitSession handles POST /sessions/{sessionID}/submit
func (h *HTTPHandler) SubmitSession(w http.ResponseWriter, r *http.Request) {
	sessionID, appErr := parseSessionID(r)
	if appErr != nil {
		utils.WriteError(w, r, appErr)
		return
	}

	var req entities.SubmitRequest
	if decErr := utils.DecodeJSON(r, &req); decErr != nil {
		utils.WriteError(w, r, decErr)
		return
	}

	resp, err := h.core.SubmitSession(r.Context(), sessionID, &req)
	if err != nil {
		utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// PurgeSession handles DELETE /sessions/{sessionID}/purge
func (h *HTTPHandler) PurgeSession(w http.ResponseWriter, r *http.Request) {
	sessionID, appErr := parseSessionID(r)
	if appErr != nil {
		utils.WriteError(w, r, appErr)
		return
	}

	if err := h.core.PurgeSession(r.Context(), sessionID); err != nil {
		utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, &StatusResponse{Status: constants.ResponseStatusPurged})
}

func parseSessionID(r *http.Request) (uuid.UUID, apperror.IError) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		return uuid.Nil, apperror.NewWithMessage(apperror.CodeBadRequest, err, apperror.MsgInvalidSessionID)
	}
	return sessionID, nil
}
