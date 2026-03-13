package entities

type SFUpsertRequest struct {
	SessionID       string            `json:"session_id"`
	ExternalIDField string            `json:"external_id_field"`
	ExternalIDValue string            `json:"external_id_value"`
	Fields          map[string]string `json:"fields"`
}

type SFUpsertResponse struct {
	RecordID  string `json:"record_id"`
	IsCreated bool   `json:"is_created"`
}

type SFUpsertAPIResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
}
