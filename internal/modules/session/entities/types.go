package entities

import (
	"time"

	"github.com/google/uuid"
)

type CallSession struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	TalkdeskCallID   string     `json:"talkdesk_call_id" db:"talkdesk_call_id"`
	AgentID          string     `json:"agent_id" db:"agent_id"`
	PatientPhone     *string    `json:"patient_phone,omitempty" db:"patient_phone"`
	Status           string     `json:"status" db:"status"`
	ParentSessionID  *uuid.UUID `json:"parent_session_id,omitempty" db:"parent_session_id"`
	SFRecordID       *string    `json:"sf_record_id,omitempty" db:"sf_record_id"`
	LanguageDetected *string    `json:"language_detected,omitempty" db:"language_detected"`
	StartedAt        time.Time  `json:"started_at" db:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty" db:"submitted_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type ExtractedField struct {
	ID            uuid.UUID `json:"id" db:"id"`
	SessionID     uuid.UUID `json:"session_id" db:"session_id"`
	FieldName     string    `json:"field_name" db:"field_name"`
	FieldValue    string    `json:"field_value" db:"field_value"`
	Confidence    float64   `json:"confidence" db:"confidence"`
	Source        string    `json:"source" db:"source"`
	Version       int       `json:"version" db:"version"`
	PreviousValue string    `json:"previous_value,omitempty" db:"previous_value"`
	TranscriptRef string    `json:"transcript_ref,omitempty" db:"transcript_ref"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type AgentOverride struct {
	ID         uuid.UUID `json:"id" db:"id"`
	SessionID  uuid.UUID `json:"session_id" db:"session_id"`
	FieldName  string    `json:"field_name" db:"field_name"`
	AIValue    string    `json:"ai_value" db:"ai_value"`
	AgentValue string    `json:"agent_value" db:"agent_value"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type SessionState struct {
	SessionID     uuid.UUID              `json:"session_id"`
	Status        string                 `json:"status"`
	Fields        map[string]*FieldState `json:"fields"`
	TranscriptLen int                    `json:"transcript_len"`
}

type FieldState struct {
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
	Version    int     `json:"version"`
}
