package entities

import "github.com/google/uuid"

type StartSessionRequest struct {
	TalkdeskCallID  string `json:"talkdesk_call_id"`
	AgentID         string `json:"agent_id"`
	PatientPhone    string `json:"patient_phone,omitempty"`
	TransferType    string `json:"transfer_type,omitempty"`
	ParentSessionID string `json:"parent_session_id,omitempty"`
}

type UpdateFieldsRequest struct {
	Overrides []FieldOverride `json:"overrides"`
}

type FieldOverride struct {
	FieldName string `json:"field_name"`
	Value     string `json:"value"`
}

type SubmitRequest struct {
	Overrides []FieldOverride `json:"overrides,omitempty"`
}

type UpdateFieldParams struct {
	SessionID  uuid.UUID
	FieldName  string
	Value      string
	Confidence float64
	Source     string
}
