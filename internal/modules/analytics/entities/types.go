package entities

import "time"

const (
	RouteAnalyticsOverview          = "/analytics/overview"
	RouteAnalyticsConditions        = "/analytics/conditions"
	RouteAnalyticsAgentPerformance  = "/analytics/agents/{agentID}/performance"
	RouteAnalyticsSentiment         = "/analytics/sentiment"

	MaxQueryRangeDays        = 90
	MinCallsForAgentMetrics  = 20
	DefaultLimit             = 20
)

type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type OverviewResponse struct {
	TimeRange           TimeRange            `json:"time_range"`
	TotalCalls          int                  `json:"total_calls"`
	AvgCallDurationMin  float64              `json:"avg_call_duration_min"`
	FieldsAutoFilledPct float64              `json:"fields_auto_filled_pct"`
	FieldsOverriddenPct float64              `json:"fields_overridden_pct"`
	AvgConfidence       float64              `json:"avg_confidence"`
	TopConditions       []ConditionCount     `json:"top_conditions"`
	TriageDistribution  map[string]int       `json:"triage_distribution"`
	SentimentSummary    map[string]int       `json:"sentiment_summary"`
}

type ConditionCount struct {
	Condition string  `json:"condition"`
	Count     int     `json:"count"`
	Pct       float64 `json:"pct"`
}

type AgentPerformanceResponse struct {
	AgentID              string            `json:"agent_id"`
	TotalCalls           int               `json:"total_calls"`
	AvgFieldsAutoFilled  float64           `json:"avg_fields_auto_filled"`
	AvgFieldsOverridden  float64           `json:"avg_fields_overridden"`
	AccuracyRate         float64           `json:"accuracy_rate"`
	MostOverriddenFields []FieldOverrideCount `json:"most_overridden_fields"`
	AvgReviewTimeSec     float64           `json:"avg_review_time_seconds"`
}

type FieldOverrideCount struct {
	Field         string `json:"field"`
	OverrideCount int    `json:"override_count"`
}

type SentimentTrendResponse struct {
	TimeRange   TimeRange         `json:"time_range"`
	Granularity string            `json:"granularity"`
	DataPoints  []SentimentPoint  `json:"data_points"`
}

type SentimentPoint struct {
	Period    string         `json:"period"`
	Emotions map[string]int `json:"emotions"`
}

type ConditionsResponse struct {
	TimeRange  TimeRange        `json:"time_range"`
	Conditions []ConditionCount `json:"conditions"`
	Total      int              `json:"total"`
}
