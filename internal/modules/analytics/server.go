package analytics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/modules/analytics/entities"
	"github.com/go-chi/chi/v5"
)

type HTTPHandler struct {
	core ICore
}

func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(entities.RouteAnalyticsOverview, h.GetOverview)
	r.Get(entities.RouteAnalyticsConditions, h.GetConditions)
	r.Get(entities.RouteAnalyticsAgentPerformance, h.GetAgentPerformance)
	r.Get(entities.RouteAnalyticsSentiment, h.GetSentimentTrend)
}

func (h *HTTPHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseTimeRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.core.GetOverview(r.Context(), from, to)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) GetConditions(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseTimeRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit := entities.DefaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	resp, err := h.core.GetConditions(r.Context(), from, to, limit)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) GetAgentPerformance(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		h.writeError(w, http.StatusBadRequest, "agent ID is required")
		return
	}

	from, to, err := parseTimeRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.core.GetAgentPerformance(r.Context(), agentID, from, to)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) GetSentimentTrend(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseTimeRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = "daily"
	}

	resp, err := h.core.GetSentimentTrend(r.Context(), from, to, granularity)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func parseTimeRange(r *http.Request) (time.Time, time.Time, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" {
		fromStr = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if toStr == "" {
		toStr = time.Now().Format("2006-01-02")
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	to = to.Add(24*time.Hour - time.Nanosecond)

	if to.Sub(from).Hours()/24 > float64(entities.MaxQueryRangeDays) {
		from = to.AddDate(0, 0, -entities.MaxQueryRangeDays)
	}

	return from, to, nil
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
