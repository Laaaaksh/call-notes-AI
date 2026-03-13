package entities

import (
	"time"

	"github.com/google/uuid"
)

type UrgencyLevel string

const (
	UrgencyLow      UrgencyLevel = "LOW"
	UrgencyMedium   UrgencyLevel = "MEDIUM"
	UrgencyHigh     UrgencyLevel = "HIGH"
	UrgencyCritical UrgencyLevel = "CRITICAL"
)

const (
	RouteSessionTriage = "/sessions/{sessionID}/triage"

	ScoreThresholdMedium   = 4
	ScoreThresholdHigh     = 7
	ScoreThresholdCritical = 9

	ModifierElderly              = 1
	ModifierPediatric            = 1
	ModifierAcuteOnset           = 1
	ModifierChronicEscalation    = 1
	ModifierMultipleSymptoms     = 1
	ModifierHighDistress         = 1
	ModifierPregnancy            = 2
	ModifierPreviousCritical     = 1
	ModifierResolvedSymptomPenalty = -3

	MultipleSymptomThreshold = 3
	ElderlyAgeThreshold      = 60
	PediatricAgeThreshold    = 5
)

type TriageAssessment struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	SessionID       uuid.UUID    `json:"session_id" db:"session_id"`
	UrgencyLevel    UrgencyLevel `json:"urgency_level" db:"urgency_level"`
	CompositeScore  int          `json:"composite_score" db:"composite_score"`
	Symptoms        []SymptomScore `json:"symptoms"`
	RedFlags        []string     `json:"red_flags" db:"red_flags"`
	ModifiersApplied []ModifierEntry `json:"modifiers_applied"`
	Version         int          `json:"version" db:"version"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" db:"updated_at"`
}

type SymptomScore struct {
	Symptom   string `json:"symptom"`
	BaseScore int    `json:"base_score"`
	IsRedFlag bool   `json:"is_red_flag"`
}

type ModifierEntry struct {
	Modifier   string `json:"modifier"`
	Adjustment int    `json:"adjustment"`
	Reason     string `json:"reason"`
}

type TriageInput struct {
	Symptom           string  `json:"symptom"`
	PatientAge        int     `json:"patient_age,omitempty"`
	DurationDays      int     `json:"duration_days,omitempty"`
	IsResolved        bool    `json:"is_resolved,omitempty"`
	SentimentIntensity float64 `json:"sentiment_intensity,omitempty"`
	PregnancyMentioned bool   `json:"pregnancy_mentioned,omitempty"`
	SymptomCount      int     `json:"symptom_count,omitempty"`
}

type TriageResponse struct {
	SessionID      string         `json:"session_id"`
	UrgencyLevel   UrgencyLevel   `json:"urgency_level"`
	CompositeScore int            `json:"composite_score"`
	Symptoms       []SymptomScore `json:"symptoms"`
	RedFlags       []string       `json:"red_flags"`
	Modifiers      []ModifierEntry `json:"modifiers"`
}
