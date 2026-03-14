package analytics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/call-notes-ai-service/internal/modules/analytics/entities"
	"github.com/call-notes-ai-service/internal/utils"
	"github.com/call-notes-ai-service/pkg/apperror"
	"github.com/go-chi/chi/v5"
)

const (
	defaultGranularity = "daily"
	dateFormat         = "2006-01-02"
	hourlyDateFormat   = "2006-01-02T15:00"
	defaultRangeDays   = 30
)

// HTTPHandler handles analytics HTTP requests
type HTTPHandler struct {
	core ICore
}

// NewHTTPHandler creates a new analytics HTTP handler
func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

// RegisterRoutes registers analytics routes on the router
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(entities.RouteAnalyticsOverview, h.GetOverview)
	r.Get(entities.RouteAnalyticsConditions, h.GetConditions)
	r.Get(entities.RouteAnalyticsAgentPerformance, h.GetAgentPerformance)
	r.Get(entities.RouteAnalyticsSentiment, h.GetSentimentTrend)
}

// GetOverview handles GET /analytics/overview
func (h *HTTPHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseTimeRange(r)
	if err != nil {
		utils.WriteError(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, err, apperror.MsgInvalidRequest))
		return
	}

	resp, err := h.core.GetOverview(r.Context(), from, to)
	if err != nil {
		utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetConditions handles GET /analytics/conditions
func (h *HTTPHandler) GetConditions(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseTimeRange(r)
	if err != nil {
		utils.WriteError(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, err, apperror.MsgInvalidRequest))
		return
	}

	limit := entities.DefaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, parseErr := strconv.Atoi(l); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}

	resp, err := h.core.GetConditions(r.Context(), from, to, limit)
	if err != nil {
		utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetAgentPerformance handles GET /analytics/agents/{agentID}
func (h *HTTPHandler) GetAgentPerformance(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		utils.WriteError(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, nil, apperror.MsgInvalidRequest))
		return
	}

	from, to, err := parseTimeRange(r)
	if err != nil {
		utils.WriteError(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, err, apperror.MsgInvalidRequest))
		return
	}

	resp, err := h.core.GetAgentPerformance(r.Context(), agentID, from, to)
	if err != nil {
		utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetSentimentTrend handles GET /analytics/sentiment
func (h *HTTPHandler) GetSentimentTrend(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseTimeRange(r)
	if err != nil {
		utils.WriteError(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, err, apperror.MsgInvalidRequest))
		return
	}

	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = defaultGranularity
	}

	resp, err := h.core.GetSentimentTrend(r.Context(), from, to, granularity)
	if err != nil {
		utils.WriteError(w, r, apperror.New(apperror.CodeInternalError, err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func parseTimeRange(r *http.Request) (time.Time, time.Time, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" {
		fromStr = time.Now().AddDate(0, 0, -defaultRangeDays).Format(dateFormat)
	}
	if toStr == "" {
		toStr = time.Now().Format(dateFormat)
	}

	from, err := time.Parse(dateFormat, fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	to, err := time.Parse(dateFormat, toStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// Include the entire "to" day
	to = to.Add(24*time.Hour - time.Nanosecond)

	if to.Sub(from).Hours()/24 > float64(entities.MaxQueryRangeDays) {
		from = to.AddDate(0, 0, -entities.MaxQueryRangeDays)
	}

	return from, to, nil
}
