// Package utils provides shared helper functions used across modules.
package utils

import (
	"encoding/json"
	"net/http"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/pkg/apperror"
	"github.com/go-chi/chi/v5/middleware"
)

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(status)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", constants.LogKeyError, err)
	}
}

// WriteError writes a standardized error response from an apperror.IError
func WriteError(w http.ResponseWriter, r *http.Request, appErr apperror.IError) {
	requestID := middleware.GetReqID(r.Context())

	resp := apperror.ErrorResponse{
		Error:     appErr.PublicMessage(),
		Code:      appErr.Code().String(),
		RequestID: requestID,
		Details:   appErr.Fields(),
	}

	WriteJSON(w, appErr.HTTPStatus(), resp)
}

// WriteErrorWithStatus writes a simple error response with a custom status code
func WriteErrorWithStatus(w http.ResponseWriter, r *http.Request, status int, message string) {
	requestID := middleware.GetReqID(r.Context())

	resp := apperror.ErrorResponse{
		Error:     message,
		RequestID: requestID,
	}

	WriteJSON(w, status, resp)
}

// DecodeJSON decodes a JSON request body into the target struct.
// Returns an apperror.IError if decoding fails.
func DecodeJSON(r *http.Request, target interface{}) apperror.IError {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return apperror.NewWithMessage(
			apperror.CodeBadRequest,
			err,
			apperror.MsgInvalidJSONBody,
		)
	}
	return nil
}
