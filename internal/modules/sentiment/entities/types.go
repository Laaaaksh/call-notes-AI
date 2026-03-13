package entities

import (
	"time"

	"github.com/google/uuid"
)

type EmotionType string

const (
	EmotionCalm      EmotionType = "calm"
	EmotionDistressed EmotionType = "distressed"
	EmotionAngry     EmotionType = "angry"
	EmotionConfused  EmotionType = "confused"
	EmotionSad       EmotionType = "sad"
)

type IntensityLevel string

const (
	IntensityLow      IntensityLevel = "LOW"
	IntensityMedium   IntensityLevel = "MEDIUM"
	IntensityHigh     IntensityLevel = "HIGH"
	IntensityCritical IntensityLevel = "CRITICAL"
)

const (
	LexiconWeight = 0.50
	PatternWeight = 0.30
	ContextWeight = 0.20

	ThresholdLow      = 0.30
	ThresholdMedium   = 0.60
	ThresholdHigh     = 0.80

	MinSTTConfidenceForSentiment = 0.50
)

type SentimentLog struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	SessionID    uuid.UUID   `json:"session_id" db:"session_id"`
	EmotionType  EmotionType `json:"emotion_type" db:"emotion_type"`
	Intensity    float64     `json:"intensity" db:"intensity"`
	LexiconScore float64    `json:"lexicon_score" db:"lexicon_score"`
	PatternScore float64    `json:"pattern_score" db:"pattern_score"`
	TriggerText  string     `json:"trigger_text" db:"trigger_text"`
	Speaker      string     `json:"speaker" db:"speaker"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type SentimentResult struct {
	Emotion        EmotionType    `json:"emotion"`
	Intensity      float64        `json:"intensity"`
	IntensityLevel IntensityLevel `json:"intensity_level"`
	LexiconScore   float64        `json:"lexicon_score"`
	PatternScore   float64        `json:"pattern_score"`
	ContextScore   float64        `json:"context_score"`
	TriggerText    string         `json:"trigger_text"`
	ShouldAlert    bool           `json:"should_alert"`
}

type SentimentSummary struct {
	DominantEmotion EmotionType     `json:"dominant_emotion"`
	AvgIntensity    float64         `json:"avg_intensity"`
	AlertsTriggered int             `json:"alerts_triggered"`
	Timeline        []TimelineEntry `json:"emotion_timeline"`
}

type TimelineEntry struct {
	Minute    int         `json:"minute"`
	Emotion   EmotionType `json:"emotion"`
	Intensity float64     `json:"intensity"`
}

type TranscriptInput struct {
	Text          string  `json:"text"`
	Speaker       string  `json:"speaker"`
	STTConfidence float64 `json:"stt_confidence"`
	WordsPerMin   float64 `json:"words_per_min"`
	PauseDuration float64 `json:"pause_duration_sec"`
	SegmentIndex  int     `json:"segment_index"`
}
