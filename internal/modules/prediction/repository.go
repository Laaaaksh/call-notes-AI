package prediction

import (
	"context"
	"time"

	"github.com/call-notes-ai-service/internal/modules/prediction/entities"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IRepository interface {
	GetPatientHistory(ctx context.Context, phone string) ([]entities.PatientHistoryEntry, error)
	UpsertHistoryEntry(ctx context.Context, entry *entities.PatientHistoryEntry) error
	GetSessionCountByPhone(ctx context.Context, phone string) (int, error)
	PurgePatientHistory(ctx context.Context, phone string) error
}

type Repository struct {
	pool *pgxpool.Pool
}

var _ IRepository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const (
	queryGetPatientHistory = `
		SELECT id, patient_phone, field_name, field_value, last_session_id,
		       occurrence_count, first_seen_at, last_seen_at
		FROM patient_history_cache
		WHERE patient_phone = $1
		ORDER BY last_seen_at DESC`

	queryUpsertHistoryEntry = `
		INSERT INTO patient_history_cache (id, patient_phone, field_name, field_value,
		    last_session_id, occurrence_count, first_seen_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (patient_phone, field_name) DO UPDATE SET
		    field_value = EXCLUDED.field_value,
		    last_session_id = EXCLUDED.last_session_id,
		    occurrence_count = patient_history_cache.occurrence_count + 1,
		    last_seen_at = EXCLUDED.last_seen_at`

	queryGetSessionCountByPhone = `
		SELECT COUNT(*) FROM call_sessions
		WHERE patient_phone = $1 AND status IN ('COMPLETED', 'SUBMITTED')`

	queryPurgePatientHistory = `DELETE FROM patient_history_cache WHERE patient_phone = $1`
)

func (r *Repository) GetPatientHistory(ctx context.Context, phone string) ([]entities.PatientHistoryEntry, error) {
	rows, err := r.pool.Query(ctx, queryGetPatientHistory, phone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []entities.PatientHistoryEntry
	for rows.Next() {
		var e entities.PatientHistoryEntry
		if err := rows.Scan(
			&e.ID, &e.PatientPhone, &e.FieldName, &e.FieldValue,
			&e.LastSessionID, &e.OccurrenceCount, &e.FirstSeenAt, &e.LastSeenAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (r *Repository) UpsertHistoryEntry(ctx context.Context, entry *entities.PatientHistoryEntry) error {
	_, err := r.pool.Exec(ctx, queryUpsertHistoryEntry,
		entry.ID, entry.PatientPhone, entry.FieldName, entry.FieldValue,
		entry.LastSessionID, entry.OccurrenceCount, entry.FirstSeenAt, entry.LastSeenAt,
	)
	return err
}

func (r *Repository) GetSessionCountByPhone(ctx context.Context, phone string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, queryGetSessionCountByPhone, phone).Scan(&count)
	return count, err
}

func (r *Repository) PurgePatientHistory(ctx context.Context, phone string) error {
	_, err := r.pool.Exec(ctx, queryPurgePatientHistory, phone)
	return err
}

func (r *Repository) UpdateHistoryFromSession(ctx context.Context, phone string, sessionID uuid.UUID, fields map[string]string) error {
	now := time.Now().UTC()
	for fieldName, fieldValue := range fields {
		entry := &entities.PatientHistoryEntry{
			ID:              uuid.New(),
			PatientPhone:    phone,
			FieldName:       fieldName,
			FieldValue:      fieldValue,
			LastSessionID:   sessionID,
			OccurrenceCount: 1,
			FirstSeenAt:     now,
			LastSeenAt:      now,
		}
		if err := r.UpsertHistoryEntry(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}
