package followup

import (
	"context"

	"github.com/call-notes-ai-service/internal/modules/followup/entities"
	"github.com/call-notes-ai-service/pkg/database"
	"github.com/google/uuid"
)

// IRepository defines the data access interface for follow-ups
type IRepository interface {
	CreateFollowUp(ctx context.Context, fu *entities.FollowUp) error
	GetFollowUps(ctx context.Context, sessionID uuid.UUID) ([]entities.FollowUp, error)
	GetFollowUp(ctx context.Context, id uuid.UUID) (*entities.FollowUp, error)
	UpdateFollowUpStatus(ctx context.Context, id uuid.UUID, status entities.FollowUpStatus, confirmedBy *string) error
}

// Repository implements IRepository using database.IPool
type Repository struct {
	pool database.IPool
}

var _ IRepository = (*Repository)(nil)

// NewRepository creates a new follow-up repository
func NewRepository(pool database.IPool) *Repository {
	return &Repository{pool: pool}
}

const (
	queryInsertFollowUp = `
		INSERT INTO follow_ups (id, session_id, follow_up_type, description,
		    raw_text, due_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	queryGetFollowUps = `
		SELECT id, session_id, follow_up_type, description, raw_text,
		       due_date, status, sf_task_id, confirmed_by, created_at, updated_at
		FROM follow_ups
		WHERE session_id = $1
		ORDER BY created_at ASC`

	queryGetFollowUp = `
		SELECT id, session_id, follow_up_type, description, raw_text,
		       due_date, status, sf_task_id, confirmed_by, created_at, updated_at
		FROM follow_ups
		WHERE id = $1`

	queryUpdateFollowUpStatus = `
		UPDATE follow_ups
		SET status = $2, confirmed_by = $3, updated_at = NOW()
		WHERE id = $1`
)

func (r *Repository) CreateFollowUp(ctx context.Context, fu *entities.FollowUp) error {
	_, err := r.pool.Exec(ctx, queryInsertFollowUp,
		fu.ID, fu.SessionID, string(fu.FollowUpType), fu.Description,
		fu.RawText, fu.DueDate, string(fu.Status), fu.CreatedAt, fu.UpdatedAt,
	)
	return err
}

func (r *Repository) GetFollowUps(ctx context.Context, sessionID uuid.UUID) ([]entities.FollowUp, error) {
	rows, err := r.pool.Query(ctx, queryGetFollowUps, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followups []entities.FollowUp
	for rows.Next() {
		var fu entities.FollowUp
		var typeStr, statusStr string
		if err := rows.Scan(
			&fu.ID, &fu.SessionID, &typeStr, &fu.Description, &fu.RawText,
			&fu.DueDate, &statusStr, &fu.SFTaskID, &fu.ConfirmedBy,
			&fu.CreatedAt, &fu.UpdatedAt,
		); err != nil {
			return nil, err
		}
		fu.FollowUpType = entities.FollowUpType(typeStr)
		fu.Status = entities.FollowUpStatus(statusStr)
		followups = append(followups, fu)
	}
	return followups, nil
}

func (r *Repository) GetFollowUp(ctx context.Context, id uuid.UUID) (*entities.FollowUp, error) {
	var fu entities.FollowUp
	var typeStr, statusStr string
	err := r.pool.QueryRow(ctx, queryGetFollowUp, id).Scan(
		&fu.ID, &fu.SessionID, &typeStr, &fu.Description, &fu.RawText,
		&fu.DueDate, &statusStr, &fu.SFTaskID, &fu.ConfirmedBy,
		&fu.CreatedAt, &fu.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	fu.FollowUpType = entities.FollowUpType(typeStr)
	fu.Status = entities.FollowUpStatus(statusStr)
	return &fu, nil
}

func (r *Repository) UpdateFollowUpStatus(ctx context.Context, id uuid.UUID, status entities.FollowUpStatus, confirmedBy *string) error {
	_, err := r.pool.Exec(ctx, queryUpdateFollowUpStatus, id, string(status), confirmedBy)
	return err
}
