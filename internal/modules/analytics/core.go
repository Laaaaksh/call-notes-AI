package analytics

import (
	"context"
	"math"
	"time"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/analytics/entities"
)

type ICore interface {
	GetOverview(ctx context.Context, from, to time.Time) (*entities.OverviewResponse, error)
	GetConditions(ctx context.Context, from, to time.Time, limit int) (*entities.ConditionsResponse, error)
	GetAgentPerformance(ctx context.Context, agentID string, from, to time.Time) (*entities.AgentPerformanceResponse, error)
	GetSentimentTrend(ctx context.Context, from, to time.Time, granularity string) (*entities.SentimentTrendResponse, error)
}

type Core struct {
	repo IRepository
}

var _ ICore = (*Core)(nil)

func NewCore(repo IRepository) ICore {
	return &Core{repo: repo}
}

func (c *Core) GetOverview(ctx context.Context, from, to time.Time) (*entities.OverviewResponse, error) {
	totalCalls, err := c.repo.GetCallCount(ctx, from, to)
	if err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgAnalyticsQueryFailed, constants.LogKeyError, err)
		return nil, err
	}

	avgDuration, _ := c.repo.GetAvgCallDuration(ctx, from, to)
	totalFields, overriddenFields, avgConfidence, _ := c.repo.GetFieldStats(ctx, from, to)
	topConditions, _ := c.repo.GetTopConditions(ctx, from, to, 10)
	triageDist, _ := c.repo.GetTriageDistribution(ctx, from, to)
	sentimentDist, _ := c.repo.GetSentimentDistribution(ctx, from, to)

	autoFilledPct := 0.0
	overriddenPct := 0.0
	if totalFields > 0 {
		autoFilledPct = math.Round((1.0-float64(overriddenFields)/float64(totalFields))*1000) / 10
		overriddenPct = math.Round(float64(overriddenFields)/float64(totalFields)*1000) / 10
	}

	for i := range topConditions {
		if totalCalls > 0 {
			topConditions[i].Pct = math.Round(float64(topConditions[i].Count)/float64(totalCalls)*1000) / 10
		}
	}

	if triageDist == nil {
		triageDist = make(map[string]int)
	}
	if sentimentDist == nil {
		sentimentDist = make(map[string]int)
	}

	return &entities.OverviewResponse{
		TimeRange:           entities.TimeRange{From: from, To: to},
		TotalCalls:          totalCalls,
		AvgCallDurationMin:  math.Round(avgDuration*10) / 10,
		FieldsAutoFilledPct: autoFilledPct,
		FieldsOverriddenPct: overriddenPct,
		AvgConfidence:       math.Round(avgConfidence*1000) / 1000,
		TopConditions:       topConditions,
		TriageDistribution:  triageDist,
		SentimentSummary:    sentimentDist,
	}, nil
}

func (c *Core) GetConditions(ctx context.Context, from, to time.Time, limit int) (*entities.ConditionsResponse, error) {
	if limit <= 0 || limit > entities.DefaultLimit {
		limit = entities.DefaultLimit
	}

	conditions, err := c.repo.GetTopConditions(ctx, from, to, limit)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, cond := range conditions {
		total += cond.Count
	}

	return &entities.ConditionsResponse{
		TimeRange:  entities.TimeRange{From: from, To: to},
		Conditions: conditions,
		Total:      total,
	}, nil
}

func (c *Core) GetAgentPerformance(ctx context.Context, agentID string, from, to time.Time) (*entities.AgentPerformanceResponse, error) {
	callCount, err := c.repo.GetAgentCallCount(ctx, agentID, from, to)
	if err != nil {
		return nil, err
	}

	if callCount < entities.MinCallsForAgentMetrics {
		return &entities.AgentPerformanceResponse{
			AgentID:    agentID,
			TotalCalls: callCount,
		}, nil
	}

	avgAutoFilled, avgOverridden, accuracy, _ := c.repo.GetAgentFieldStats(ctx, agentID, from, to)
	topOverrides, _ := c.repo.GetAgentTopOverrides(ctx, agentID, from, to, 10)

	return &entities.AgentPerformanceResponse{
		AgentID:              agentID,
		TotalCalls:           callCount,
		AvgFieldsAutoFilled:  math.Round(avgAutoFilled*10) / 10,
		AvgFieldsOverridden:  math.Round(avgOverridden*10) / 10,
		AccuracyRate:         math.Round(accuracy*1000) / 1000,
		MostOverriddenFields: topOverrides,
	}, nil
}

func (c *Core) GetSentimentTrend(ctx context.Context, from, to time.Time, granularity string) (*entities.SentimentTrendResponse, error) {
	if granularity == "" {
		granularity = "daily"
	}

	points, err := c.repo.GetSentimentTrend(ctx, from, to, granularity)
	if err != nil {
		return nil, err
	}

	return &entities.SentimentTrendResponse{
		TimeRange:   entities.TimeRange{From: from, To: to},
		Granularity: granularity,
		DataPoints:  points,
	}, nil
}
