package entities

type EntityType string

const (
	EntitySymptom      EntityType = "symptom"
	EntityBodyPart     EntityType = "body_part"
	EntityCondition    EntityType = "condition"
	EntityMedication   EntityType = "medication"
	EntityDuration     EntityType = "duration"
	EntitySeverity     EntityType = "severity"
	EntityAge          EntityType = "age"
	EntityName         EntityType = "name"
	EntityPhone        EntityType = "phone"
	EntityGender       EntityType = "gender"
	EntityAllergy      EntityType = "allergy"
	EntityFollowUp     EntityType = "follow_up"
	EntityICD10        EntityType = "icd10_code"
	EntityNegated      EntityType = "negated"
)

type MedicalEntity struct {
	Type           EntityType `json:"type"`
	RawValue       string     `json:"raw_value"`
	NormalizedValue string    `json:"normalized_value"`
	Confidence     float64    `json:"confidence"`
	SourceLayer    string     `json:"source_layer"`
	TranscriptRef  string     `json:"transcript_ref"`
	IsNegated      bool       `json:"is_negated"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type ExtractionResult struct {
	Entities []MedicalEntity `json:"entities"`
	SessionID string        `json:"session_id"`
	SegmentID string        `json:"segment_id"`
}

type TranscriptSegment struct {
	SessionID  string  `json:"session_id"`
	Text       string  `json:"text"`
	Speaker    string  `json:"speaker"`
	Confidence float64 `json:"confidence"`
	IsFinal    bool    `json:"is_final"`
	Sequence   int     `json:"sequence"`
	Timestamp  float64 `json:"timestamp"`
}

const (
	ErrMsgExtractionFailed  = "entity extraction failed"
	ErrMsgNERUnavailable    = "medical NER service unavailable"
	ErrMsgLLMUnavailable    = "LLM reasoning service unavailable"
)
