package prediction

import (
	"net/http"

	"github.com/call-notes-ai-service/internal/modules/prediction/entities"
	"github.com/call-notes-ai-service/internal/utils"
	"github.com/call-notes-ai-service/pkg/apperror"
	"github.com/go-chi/chi/v5"
)

// HTTPHandler handles prediction HTTP requests
type HTTPHandler struct {
	core ICore
}

// NewHTTPHandler creates a new prediction HTTP handler
func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

// RegisterRoutes registers prediction routes on the router
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(entities.RoutePatientHistory, h.GetPatientHistory)
}

// GetPatientHistory handles GET /patients/{phone}/history
func (h *HTTPHandler) GetPatientHistory(w http.ResponseWriter, r *http.Request) {
	phone := chi.URLParam(r, "phone")
	if phone == "" {
		utils.WriteError(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, nil, apperror.MsgPatientNotFound))
		return
	}

	resp, err := h.core.GetPredictedFields(r.Context(), phone)
	if err != nil {
		utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
