package prediction

import (
	"encoding/json"
	"net/http"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/modules/prediction/entities"
	"github.com/go-chi/chi/v5"
)

type HTTPHandler struct {
	core ICore
}

func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(entities.RoutePatientHistory, h.GetPatientHistory)
}

func (h *HTTPHandler) GetPatientHistory(w http.ResponseWriter, r *http.Request) {
	phone := chi.URLParam(r, "phone")
	if phone == "" {
		h.writeError(w, http.StatusBadRequest, "phone number is required")
		return
	}

	resp, err := h.core.GetPredictedFields(r.Context(), phone)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
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
