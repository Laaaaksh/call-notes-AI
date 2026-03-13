package entities

type SessionResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

type SessionFieldsResponse struct {
	SessionID string          `json:"session_id"`
	Fields    []FieldResponse `json:"fields"`
	Status    string          `json:"status"`
}

type FieldResponse struct {
	FieldName  string  `json:"field_name"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

type SubmitResponse struct {
	SessionID  string `json:"session_id"`
	SFRecordID string `json:"sf_record_id"`
	Status     string `json:"status"`
}
