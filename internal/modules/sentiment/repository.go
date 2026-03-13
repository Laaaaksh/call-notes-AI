package sentiment

import (
	"context"

	"github.com/call-notes-ai-service/internal/modules/sentiment/entities"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IRepository interface {
	CreateSentimentLog(ctx context.Context, log *entities.SentimentLog) error
	GetSentimentLogs(ctx context.Context, sessionID uuid.UUID) ([]entities.SentimentLog, error)
}

type Repository struct {
	pool *pgxpool.Pool
}

var _ IRepository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const (
	queryInsertSentimentLog = `
		INSERT INTO sentiment_logs (id, session_id, emotion_type, intensity,
		    lexicon_score, pattern_score, trigger_text, speaker, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`

	queryGetSentimentLogs = `
		SELECT id, session_id, emotion_type, intensity, lexicon_score,
		       pattern_score, trigger_text, speaker, created_at
		FROM sentiment_logs
		WHERE session_id = $1
		ORDER BY created_at ASC`
)

func (r *Repository) CreateSentimentLog(ctx context.Context, log *entities.SentimentLog) error {
	_, err := r.pool.Exec(ctx, queryInsertSentimentLog,
		log.ID, log.SessionID, string(log.EmotionType), log.Intensity,
		log.LexiconScore, log.PatternScore, log.TriggerText, log.Speaker,
	)
	return err
}

func (r *Repository) GetSentimentLogs(ctx context.Context, sessionID uuid.UUID) ([]entities.SentimentLog, error) {
	rows, err := r.pool.Query(ctx, queryGetSentimentLogs, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []entities.SentimentLog
	for rows.Next() {
		var l entities.SentimentLog
		var emotionStr string
		if err := rows.Scan(
			&l.ID, &l.SessionID, &emotionStr, &l.Intensity,
			&l.LexiconScore, &l.PatternScore, &l.TriggerText,
			&l.Speaker, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		l.EmotionType = entities.EmotionType(emotionStr)
		logs = append(logs, l)
	}
	return logs, nil
}
