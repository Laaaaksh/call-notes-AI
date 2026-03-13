package entities

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoutePatientHistory = "/patients/{phone}/history"

	SourceHistory = "history"

	BaseHistoryConfidence    = 0.90
	MonthlyDecayRate         = 0.03
	MinDecayFactor           = 0.50
	RepetitionBoostPerOccurrence = 0.03
	MaxRepetitionBoost       = 1.15
	MinPreFillConfidence     = 0.60
	MaxHistorySessions       = 5
)

type PatientHistoryEntry struct {
	ID              uuid.UUID `json:"id" db:"id"`
	PatientPhone    string    `json:"patient_phone" db:"patient_phone"`
	FieldName       string    `json:"field_name" db:"field_name"`
	FieldValue      string    `json:"field_value" db:"field_value"`
	LastSessionID   uuid.UUID `json:"last_session_id" db:"last_session_id"`
	OccurrenceCount int       `json:"occurrence_count" db:"occurrence_count"`
	FirstSeenAt     time.Time `json:"first_seen_at" db:"first_seen_at"`
	LastSeenAt      time.Time `json:"last_seen_at" db:"last_seen_at"`
}

type PredictedField struct {
	FieldName       string    `json:"field_name"`
	FieldValue      string    `json:"field_value"`
	Confidence      float64   `json:"confidence"`
	Source          string    `json:"source"`
	OccurrenceCount int       `json:"occurrence_count"`
	LastSeenAt      time.Time `json:"last_seen_at"`
}

type PatientHistoryResponse struct {
	PatientPhone    string            `json:"patient_phone"`
	TotalSessions   int               `json:"total_sessions"`
	PredictedFields []PredictedField  `json:"predicted_fields"`
	LastVisitDate   *time.Time        `json:"last_visit_date,omitempty"`
}
