package triage

import (
	"context"
	"encoding/json"

	"github.com/call-notes-ai-service/internal/modules/triage/entities"
	"github.com/call-notes-ai-service/pkg/database"
	"github.com/google/uuid"
)

// IRepository defines the data access interface for triage
type IRepository interface {
	UpsertAssessment(ctx context.Context, assessment *entities.TriageAssessment) error
	GetLatestAssessment(ctx context.Context, sessionID uuid.UUID) (*entities.TriageAssessment, error)
}

// Repository implements IRepository using database.IPool
type Repository struct {
	pool database.IPool
}

var _ IRepository = (*Repository)(nil)

// NewRepository creates a new triage repository
func NewRepository(pool database.IPool) *Repository {
	return &Repository{pool: pool}
}

const (
	queryUpsertAssessment = `
		INSERT INTO triage_assessments (id, session_id, urgency_level, composite_score,
		    symptoms, red_flags, modifiers_applied, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
		    urgency_level = EXCLUDED.urgency_level,
		    composite_score = EXCLUDED.composite_score,
		    symptoms = EXCLUDED.symptoms,
		    red_flags = EXCLUDED.red_flags,
		    modifiers_applied = EXCLUDED.modifiers_applied,
		    version = EXCLUDED.version,
		    updated_at = EXCLUDED.updated_at`

	queryGetLatestAssessment = `
		SELECT id, session_id, urgency_level, composite_score,
		       symptoms, red_flags, modifiers_applied, version,
		       created_at, updated_at
		FROM triage_assessments
		WHERE session_id = $1
		ORDER BY version DESC LIMIT 1`
)

func (r *Repository) UpsertAssessment(ctx context.Context, assessment *entities.TriageAssessment) error {
	symptomsJSON, err := json.Marshal(assessment.Symptoms)
	if err != nil {
		return err
	}
	modifiersJSON, err := json.Marshal(assessment.ModifiersApplied)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, queryUpsertAssessment,
		assessment.ID, assessment.SessionID, string(assessment.UrgencyLevel),
		assessment.CompositeScore, symptomsJSON, assessment.RedFlags,
		modifiersJSON, assessment.Version, assessment.CreatedAt, assessment.UpdatedAt,
	)
	return err
}

func (r *Repository) GetLatestAssessment(ctx context.Context, sessionID uuid.UUID) (*entities.TriageAssessment, error) {
	var a entities.TriageAssessment
	var urgencyStr string
	var symptomsJSON, modifiersJSON []byte

	err := r.pool.QueryRow(ctx, queryGetLatestAssessment, sessionID).Scan(
		&a.ID, &a.SessionID, &urgencyStr, &a.CompositeScore,
		&symptomsJSON, &a.RedFlags, &modifiersJSON, &a.Version,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	a.UrgencyLevel = entities.UrgencyLevel(urgencyStr)
	_ = json.Unmarshal(symptomsJSON, &a.Symptoms)
	_ = json.Unmarshal(modifiersJSON, &a.ModifiersApplied)

	return &a, nil
}
