package analytics

import (
	"context"
	"time"

	"github.com/call-notes-ai-service/internal/modules/analytics/entities"
	"github.com/call-notes-ai-service/pkg/database"
)

// IRepository defines the data access interface for analytics
type IRepository interface {
	GetCallCount(ctx context.Context, from, to time.Time) (int, error)
	GetAvgCallDuration(ctx context.Context, from, to time.Time) (float64, error)
	GetFieldStats(ctx context.Context, from, to time.Time) (totalFields int, overriddenFields int, avgConfidence float64, err error)
	GetTopConditions(ctx context.Context, from, to time.Time, limit int) ([]entities.ConditionCount, error)
	GetTriageDistribution(ctx context.Context, from, to time.Time) (map[string]int, error)
	GetSentimentDistribution(ctx context.Context, from, to time.Time) (map[string]int, error)
	GetAgentCallCount(ctx context.Context, agentID string, from, to time.Time) (int, error)
	GetAgentFieldStats(ctx context.Context, agentID string, from, to time.Time) (avgAutoFilled float64, avgOverridden float64, accuracy float64, err error)
	GetAgentTopOverrides(ctx context.Context, agentID string, from, to time.Time, limit int) ([]entities.FieldOverrideCount, error)
	GetSentimentTrend(ctx context.Context, from, to time.Time, granularity string) ([]entities.SentimentPoint, error)
}

// Repository implements IRepository using database.IPool
type Repository struct {
	pool database.IPool
}

var _ IRepository = (*Repository)(nil)

// NewRepository creates a new analytics repository
func NewRepository(pool database.IPool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetCallCount(ctx context.Context, from, to time.Time) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM call_sessions WHERE created_at BETWEEN $1 AND $2`,
		from, to,
	).Scan(&count)
	return count, err
}

func (r *Repository) GetAvgCallDuration(ctx context.Context, from, to time.Time) (float64, error) {
	var avg float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (ended_at - started_at)) / 60.0), 0)
		 FROM call_sessions
		 WHERE ended_at IS NOT NULL AND created_at BETWEEN $1 AND $2`,
		from, to,
	).Scan(&avg)
	return avg, err
}

func (r *Repository) GetFieldStats(ctx context.Context, from, to time.Time) (int, int, float64, error) {
	var totalFields, overriddenFields int
	var avgConfidence float64

	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(AVG(confidence), 0)
		 FROM extracted_fields ef
		 JOIN call_sessions cs ON ef.session_id = cs.id
		 WHERE cs.created_at BETWEEN $1 AND $2`,
		from, to,
	).Scan(&totalFields, &avgConfidence)
	if err != nil {
		return 0, 0, 0, err
	}

	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*)
		 FROM agent_overrides ao
		 JOIN call_sessions cs ON ao.session_id = cs.id
		 WHERE cs.created_at BETWEEN $1 AND $2`,
		from, to,
	).Scan(&overriddenFields)

	return totalFields, overriddenFields, avgConfidence, err
}

func (r *Repository) GetTopConditions(ctx context.Context, from, to time.Time, limit int) ([]entities.ConditionCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ef.field_value, COUNT(*) as cnt
		 FROM extracted_fields ef
		 JOIN call_sessions cs ON ef.session_id = cs.id
		 WHERE ef.field_name IN ('primary_symptom', 'suspected_condition')
		   AND cs.created_at BETWEEN $1 AND $2
		 GROUP BY ef.field_value
		 ORDER BY cnt DESC
		 LIMIT $3`,
		from, to, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conditions []entities.ConditionCount
	for rows.Next() {
		var c entities.ConditionCount
		if err := rows.Scan(&c.Condition, &c.Count); err != nil {
			return nil, err
		}
		conditions = append(conditions, c)
	}
	return conditions, nil
}

func (r *Repository) GetTriageDistribution(ctx context.Context, from, to time.Time) (map[string]int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ta.urgency_level, COUNT(*)
		 FROM triage_assessments ta
		 JOIN call_sessions cs ON ta.session_id = cs.id
		 WHERE cs.created_at BETWEEN $1 AND $2
		 GROUP BY ta.urgency_level`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, err
		}
		dist[level] = count
	}
	return dist, nil
}

func (r *Repository) GetSentimentDistribution(ctx context.Context, from, to time.Time) (map[string]int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT sl.emotion_type, COUNT(*)
		 FROM sentiment_logs sl
		 JOIN call_sessions cs ON sl.session_id = cs.id
		 WHERE cs.created_at BETWEEN $1 AND $2
		 GROUP BY sl.emotion_type`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var emotion string
		var count int
		if err := rows.Scan(&emotion, &count); err != nil {
			return nil, err
		}
		dist[emotion] = count
	}
	return dist, nil
}

func (r *Repository) GetAgentCallCount(ctx context.Context, agentID string, from, to time.Time) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM call_sessions
		 WHERE agent_id = $1 AND created_at BETWEEN $2 AND $3`,
		agentID, from, to,
	).Scan(&count)
	return count, err
}

func (r *Repository) GetAgentFieldStats(ctx context.Context, agentID string, from, to time.Time) (float64, float64, float64, error) {
	var avgAutoFilled, avgOverridden float64
	err := r.pool.QueryRow(ctx,
		`SELECT
		   COALESCE(AVG(field_count), 0),
		   COALESCE(AVG(override_count), 0)
		 FROM (
		   SELECT cs.id,
		     (SELECT COUNT(*) FROM extracted_fields WHERE session_id = cs.id) as field_count,
		     (SELECT COUNT(*) FROM agent_overrides WHERE session_id = cs.id) as override_count
		   FROM call_sessions cs
		   WHERE cs.agent_id = $1 AND cs.created_at BETWEEN $2 AND $3
		 ) sub`,
		agentID, from, to,
	).Scan(&avgAutoFilled, &avgOverridden)
	if err != nil {
		return 0, 0, 0, err
	}

	accuracy := 0.0
	if avgAutoFilled > 0 {
		accuracy = 1.0 - (avgOverridden / avgAutoFilled)
		if accuracy < 0 {
			accuracy = 0
		}
	}

	return avgAutoFilled, avgOverridden, accuracy, nil
}

func (r *Repository) GetAgentTopOverrides(ctx context.Context, agentID string, from, to time.Time, limit int) ([]entities.FieldOverrideCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ao.field_name, COUNT(*) as cnt
		 FROM agent_overrides ao
		 JOIN call_sessions cs ON ao.session_id = cs.id
		 WHERE cs.agent_id = $1 AND cs.created_at BETWEEN $2 AND $3
		 GROUP BY ao.field_name
		 ORDER BY cnt DESC
		 LIMIT $4`,
		agentID, from, to, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overrides []entities.FieldOverrideCount
	for rows.Next() {
		var o entities.FieldOverrideCount
		if err := rows.Scan(&o.Field, &o.OverrideCount); err != nil {
			return nil, err
		}
		overrides = append(overrides, o)
	}
	return overrides, nil
}

func (r *Repository) GetSentimentTrend(ctx context.Context, from, to time.Time, granularity string) ([]entities.SentimentPoint, error) {
	truncExpr := "day"
	if granularity == "weekly" {
		truncExpr = "week"
	} else if granularity == "hourly" {
		truncExpr = "hour"
	}

	rows, err := r.pool.Query(ctx,
		`SELECT date_trunc('`+truncExpr+`', sl.created_at) as period,
		        sl.emotion_type, COUNT(*)
		 FROM sentiment_logs sl
		 WHERE sl.created_at BETWEEN $1 AND $2
		 GROUP BY period, sl.emotion_type
		 ORDER BY period`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pointMap := make(map[string]*entities.SentimentPoint)
	for rows.Next() {
		var period time.Time
		var emotion string
		var count int
		if err := rows.Scan(&period, &emotion, &count); err != nil {
			return nil, err
		}

		key := period.Format("2006-01-02")
		if granularity == "hourly" {
			key = period.Format("2006-01-02T15:00")
		}

		if _, exists := pointMap[key]; !exists {
			pointMap[key] = &entities.SentimentPoint{
				Period:   key,
				Emotions: make(map[string]int),
			}
		}
		pointMap[key].Emotions[emotion] = count
	}

	points := make([]entities.SentimentPoint, 0, len(pointMap))
	for _, p := range pointMap {
		points = append(points, *p)
	}
	return points, nil
}
