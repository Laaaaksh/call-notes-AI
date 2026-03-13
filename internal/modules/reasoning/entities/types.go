package entities

type ReasoningRequest struct {
	SessionID      string            `json:"session_id"`
	TranscriptText string            `json:"transcript_text"`
	ExistingFields map[string]string `json:"existing_fields"`
	ConflictFields []string          `json:"conflict_fields"`
}

type ReasoningResponse struct {
	ResolvedFields map[string]string `json:"resolved_fields"`
	Corrections    []Correction      `json:"corrections"`
	CallSummary    string            `json:"call_summary"`
}

type Correction struct {
	FieldName string `json:"field_name"`
	OldValue  string `json:"old_value"`
	NewValue  string `json:"new_value"`
	Reason    string `json:"reason"`
}
