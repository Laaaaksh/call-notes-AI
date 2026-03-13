package entities

type TranscriptChunk struct {
	SessionID  string  `json:"session_id"`
	Text       string  `json:"text"`
	Speaker    string  `json:"speaker"`
	Confidence float64 `json:"confidence"`
	IsFinal    bool    `json:"is_final"`
	Sequence   int     `json:"sequence"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	Language   string  `json:"language"`
}

type FullTranscript struct {
	SessionID string            `json:"session_id"`
	Chunks    []TranscriptChunk `json:"chunks"`
}

const (
	ErrMsgTranscriptionFailed = "transcription processing failed"
)
