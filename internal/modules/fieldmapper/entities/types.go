package entities

type SalesforceFieldMapping struct {
	FieldName    string `json:"field_name"`
	SFFieldAPI   string `json:"sf_field_api"`
	DataType     string `json:"data_type"`
	Required     bool   `json:"required"`
	MaxLength    int    `json:"max_length,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
}

type MappedFields struct {
	SessionID string            `json:"session_id"`
	Fields    map[string]string `json:"fields"`
}
