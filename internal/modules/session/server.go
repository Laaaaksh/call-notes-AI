package session

import (
	"encoding/json"
	"net/http"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/modules/session/entities"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type HTTPHandler struct {
	core ICore
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Post(entities.RouteSessionsCreate, h.CreateSession)
	r.Get(entities.RouteSessionsGet, h.GetSession)
	r.Patch(entities.RouteSessionsFields, h.UpdateFields)
	r.Post(entities.RouteSessionsSubmit, h.SubmitSession)
	r.Delete(entities.RouteSessionsPurge, h.PurgeSession)
}

func (h *HTTPHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req entities.StartSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrMsgInvalidRequestBody)
		return
	}

	resp, err := h.core.StartSession(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

func (h *HTTPHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrMsgInvalidSessionID)
		return
	}

	state, err := h.core.GetSessionState(r.Context(), sessionID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, state)
}

func (h *HTTPHandler) UpdateFields(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrMsgInvalidSessionID)
		return
	}

	var req entities.UpdateFieldsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrMsgInvalidRequestBody)
		return
	}

	for _, o := range req.Overrides {
		if err := h.core.ApplyAgentOverride(r.Context(), sessionID, o.FieldName, o.Value); err != nil {
			h.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	h.writeJSON(w, http.StatusOK, &StatusResponse{Status: constants.ResponseStatusUpdated})
}

func (h *HTTPHandler) SubmitSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrMsgInvalidSessionID)
		return
	}

	var req entities.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrMsgInvalidRequestBody)
		return
	}

	resp, err := h.core.SubmitSession(r.Context(), sessionID, &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) PurgeSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, constants.ErrMsgInvalidSessionID)
		return
	}

	if err := h.core.PurgeSession(r.Context(), sessionID); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, &StatusResponse{Status: "purged"})
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
	_ = json.NewEncoder(w).Encode(&ErrorResponse{Error: message})
}
