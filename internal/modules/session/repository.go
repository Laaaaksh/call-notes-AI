package session

//go:generate mockgen -source=repository.go -destination=mock/mock_repository.go -package=mock

import (
	"context"
	"time"

	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/session/entities"
	"github.com/call-notes-ai-service/pkg/database"
	"github.com/google/uuid"
)

// IRepository defines the data access interface for session management
type IRepository interface {
	CreateSession(ctx context.Context, session *entities.CallSession) error
	GetSession(ctx context.Context, id uuid.UUID) (*entities.CallSession, error)
	UpdateSessionStatus(ctx context.Context, id uuid.UUID, status string, endedAt *time.Time) error
	UpsertField(ctx context.Context, field *entities.ExtractedField) error
	GetLatestFields(ctx context.Context, sessionID uuid.UUID) ([]entities.ExtractedField, error)
	GetFieldByName(ctx context.Context, sessionID uuid.UUID, fieldName string) (*entities.ExtractedField, error)
	CreateOverride(ctx context.Context, override *entities.AgentOverride) error
	PurgeSession(ctx context.Context, id uuid.UUID) error
}

// Repository implements IRepository using database.IPool
type Repository struct {
	pool database.IPool
}

var _ IRepository = (*Repository)(nil)

// NewRepository creates a new session repository
func NewRepository(pool database.IPool) *Repository {
	return &Repository{pool: pool}
}

const (
	queryInsertSession = `
		INSERT INTO call_sessions (id, talkdesk_call_id, agent_id, patient_phone, status, parent_session_id, started_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	queryGetSession = `SELECT id, talkdesk_call_id, agent_id, patient_phone, status, parent_session_id, sf_record_id, started_at, ended_at, submitted_at, created_at, updated_at FROM call_sessions WHERE id = $1`

	queryUpdateSessionStatus = `UPDATE call_sessions SET status = $2, ended_at = $3, updated_at = NOW() WHERE id = $1`

	queryUpsertField = `
		INSERT INTO extracted_fields (id, session_id, field_name, field_value, confidence, source, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 1, $7, $8)
		ON CONFLICT (session_id, field_name, version) DO UPDATE SET
			field_value = EXCLUDED.field_value,
			confidence = EXCLUDED.confidence,
			source = EXCLUDED.source,
			updated_at = EXCLUDED.updated_at`

	queryGetLatestFields = `
		SELECT DISTINCT ON (field_name) id, session_id, field_name, field_value, confidence, source, version, created_at, updated_at
		FROM extracted_fields WHERE session_id = $1
		ORDER BY field_name, version DESC`

	queryGetFieldByName = `
		SELECT id, session_id, field_name, field_value, confidence, source, version, created_at, updated_at
		FROM extracted_fields WHERE session_id = $1 AND field_name = $2
		ORDER BY version DESC LIMIT 1`

	queryInsertOverride = `
		INSERT INTO agent_overrides (id, session_id, field_name, ai_value, agent_value, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	queryPurgeAuditLogs   = `DELETE FROM audit_logs WHERE session_id = $1`
	queryPurgeOverrides   = `DELETE FROM agent_overrides WHERE session_id = $1`
	queryPurgeFields      = `DELETE FROM extracted_fields WHERE session_id = $1`
	queryPurgeSession     = `DELETE FROM call_sessions WHERE id = $1`
)

func (r *Repository) CreateSession(ctx context.Context, session *entities.CallSession) error {
	_, err := r.pool.Exec(ctx, queryInsertSession,
		session.ID, session.TalkdeskCallID, session.AgentID, session.PatientPhone,
		session.Status, session.ParentSessionID, session.StartedAt, session.CreatedAt, session.UpdatedAt,
	)
	if err != nil {
		logger.Ctx(ctx).Errorw("Failed to create session", "error", err)
	}
	return err
}

func (r *Repository) GetSession(ctx context.Context, id uuid.UUID) (*entities.CallSession, error) {
	var s entities.CallSession
	err := r.pool.QueryRow(ctx, queryGetSession, id).Scan(
		&s.ID, &s.TalkdeskCallID, &s.AgentID, &s.PatientPhone, &s.Status,
		&s.ParentSessionID, &s.SFRecordID, &s.StartedAt, &s.EndedAt, &s.SubmittedAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) UpdateSessionStatus(ctx context.Context, id uuid.UUID, status string, endedAt *time.Time) error {
	_, err := r.pool.Exec(ctx, queryUpdateSessionStatus, id, status, endedAt)
	return err
}

func (r *Repository) UpsertField(ctx context.Context, field *entities.ExtractedField) error {
	_, err := r.pool.Exec(ctx, queryUpsertField,
		field.ID, field.SessionID, field.FieldName, field.FieldValue,
		field.Confidence, field.Source, field.CreatedAt, field.UpdatedAt,
	)
	return err
}

func (r *Repository) GetLatestFields(ctx context.Context, sessionID uuid.UUID) ([]entities.ExtractedField, error) {
	rows, err := r.pool.Query(ctx, queryGetLatestFields, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []entities.ExtractedField
	for rows.Next() {
		var f entities.ExtractedField
		if err := rows.Scan(&f.ID, &f.SessionID, &f.FieldName, &f.FieldValue, &f.Confidence, &f.Source, &f.Version, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	return fields, nil
}

func (r *Repository) GetFieldByName(ctx context.Context, sessionID uuid.UUID, fieldName string) (*entities.ExtractedField, error) {
	var f entities.ExtractedField
	err := r.pool.QueryRow(ctx, queryGetFieldByName, sessionID, fieldName).Scan(
		&f.ID, &f.SessionID, &f.FieldName, &f.FieldValue, &f.Confidence, &f.Source, &f.Version, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *Repository) CreateOverride(ctx context.Context, override *entities.AgentOverride) error {
	_, err := r.pool.Exec(ctx, queryInsertOverride,
		override.ID, override.SessionID, override.FieldName, override.AIValue, override.AgentValue, override.CreatedAt,
	)
	return err
}

func (r *Repository) PurgeSession(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	purgeQueries := []string{queryPurgeAuditLogs, queryPurgeOverrides, queryPurgeFields, queryPurgeSession}
	for _, q := range purgeQueries {
		if _, err := tx.Exec(ctx, q, id); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
