package entities

import (
	"time"

	"github.com/google/uuid"
)

type FollowUpType string

const (
	FollowUpCallback           FollowUpType = "CALLBACK"
	FollowUpLabTest            FollowUpType = "LAB_TEST"
	FollowUpPrescriptionRefill FollowUpType = "PRESCRIPTION_REFILL"
	FollowUpAppointment        FollowUpType = "APPOINTMENT"
	FollowUpConditional        FollowUpType = "CONDITIONAL_CALLBACK"
)

type FollowUpStatus string

const (
	StatusDetected    FollowUpStatus = "DETECTED"
	StatusConfirmed   FollowUpStatus = "CONFIRMED"
	StatusCreatedInSF FollowUpStatus = "CREATED_IN_SF"
	StatusDismissed   FollowUpStatus = "DISMISSED"
)

const (
	RouteSessionFollowups        = "/sessions/{sessionID}/followups"
	RouteSessionFollowupsConfirm = "/sessions/{sessionID}/followups/confirm"

	DefaultCallbackDays          = 14
	DefaultLabTestDays           = 7
	DefaultPrescriptionDays      = 30
	DefaultAppointmentDays       = 7
	DefaultConditionalDays       = 7
)

type FollowUp struct {
	ID           uuid.UUID      `json:"id" db:"id"`
	SessionID    uuid.UUID      `json:"session_id" db:"session_id"`
	FollowUpType FollowUpType   `json:"follow_up_type" db:"follow_up_type"`
	Description  string         `json:"description" db:"description"`
	RawText      string         `json:"raw_text" db:"raw_text"`
	DueDate      *time.Time     `json:"due_date,omitempty" db:"due_date"`
	Status       FollowUpStatus `json:"status" db:"status"`
	SFTaskID     *string        `json:"sf_task_id,omitempty" db:"sf_task_id"`
	ConfirmedBy  *string        `json:"confirmed_by,omitempty" db:"confirmed_by"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

type FollowUpDetection struct {
	Type        FollowUpType `json:"type"`
	Description string       `json:"description"`
	RawText     string       `json:"raw_text"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	DurationDays int         `json:"duration_days,omitempty"`
}

type ConfirmFollowUpRequest struct {
	FollowUpID string `json:"followup_id"`
	Confirmed  bool   `json:"confirmed"`
	AgentID    string `json:"agent_id"`
}

type FollowUpResponse struct {
	FollowUps []FollowUp `json:"followups"`
}
