package sentiment

import (
	"context"
	"strings"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/sentiment/entities"
	"github.com/google/uuid"
)

type ICore interface {
	AnalyzeSegment(ctx context.Context, sessionID uuid.UUID, input *entities.TranscriptInput) (*entities.SentimentResult, error)
	GetSessionSentiment(ctx context.Context, sessionID uuid.UUID) ([]entities.SentimentLog, error)
}

type Core struct {
	repo    IRepository
	lexicon *Lexicon
}

var _ ICore = (*Core)(nil)

func NewCore(repo IRepository) ICore {
	return &Core{
		repo:    repo,
		lexicon: NewLexicon(),
	}
}

func (c *Core) AnalyzeSegment(ctx context.Context, sessionID uuid.UUID, input *entities.TranscriptInput) (*entities.SentimentResult, error) {
	if input.STTConfidence < entities.MinSTTConfidenceForSentiment {
		return &entities.SentimentResult{
			Emotion:   entities.EmotionCalm,
			Intensity: 0,
		}, nil
	}

	lexiconScores := c.lexicon.Score(input.Text)
	patternScores := analyzePatterns(input)

	bestEmotion := entities.EmotionCalm
	bestScore := 0.0

	for _, emotion := range []entities.EmotionType{
		entities.EmotionDistressed, entities.EmotionAngry,
		entities.EmotionConfused, entities.EmotionSad,
	} {
		lexScore := lexiconScores[emotion]
		patScore := patternScores[emotion]
		contextScore := 0.0

		combined := (entities.LexiconWeight * lexScore) +
			(entities.PatternWeight * patScore) +
			(entities.ContextWeight * contextScore)

		if combined > bestScore {
			bestScore = combined
			bestEmotion = emotion
		}
	}

	if bestScore < entities.ThresholdLow {
		bestEmotion = entities.EmotionCalm
		bestScore = 0.0
	}

	intensityLevel := classifyIntensity(bestScore)
	shouldAlert := bestScore >= entities.ThresholdMedium

	result := &entities.SentimentResult{
		Emotion:        bestEmotion,
		Intensity:      bestScore,
		IntensityLevel: intensityLevel,
		LexiconScore:   lexiconScores[bestEmotion],
		PatternScore:   patternScores[bestEmotion],
		ContextScore:   0.0,
		TriggerText:    input.Text,
		ShouldAlert:    shouldAlert,
	}

	if bestScore >= entities.ThresholdLow {
		logEntry := &entities.SentimentLog{
			ID:           uuid.New(),
			SessionID:    sessionID,
			EmotionType:  bestEmotion,
			Intensity:    bestScore,
			LexiconScore: lexiconScores[bestEmotion],
			PatternScore: patternScores[bestEmotion],
			TriggerText:  truncateText(input.Text, 500),
			Speaker:      input.Speaker,
		}
		if err := c.repo.CreateSentimentLog(ctx, logEntry); err != nil {
			logger.Ctx(ctx).Errorw(constants.LogMsgSentimentLogFailed,
				constants.LogKeyError, err,
				constants.LogFieldSessionID, sessionID.String(),
			)
		}
	}

	if shouldAlert {
		logger.Ctx(ctx).Infow(constants.LogMsgSentimentAlertTriggered,
			constants.LogFieldSessionID, sessionID.String(),
			constants.LogFieldEmotionType, string(bestEmotion),
			constants.LogFieldIntensity, bestScore,
		)
	}

	return result, nil
}

func (c *Core) GetSessionSentiment(ctx context.Context, sessionID uuid.UUID) ([]entities.SentimentLog, error) {
	return c.repo.GetSentimentLogs(ctx, sessionID)
}

func analyzePatterns(input *entities.TranscriptInput) map[entities.EmotionType]float64 {
	scores := make(map[entities.EmotionType]float64)

	if input.WordsPerMin > 200 {
		scores[entities.EmotionAngry] += 0.6
	} else if input.WordsPerMin > 170 {
		scores[entities.EmotionAngry] += 0.3
	}

	if input.WordsPerMin < 80 && input.WordsPerMin > 0 {
		scores[entities.EmotionDistressed] += 0.4
		scores[entities.EmotionSad] += 0.3
	}

	if input.PauseDuration > 3.0 {
		scores[entities.EmotionDistressed] += 0.5
	} else if input.PauseDuration > 2.0 {
		scores[entities.EmotionDistressed] += 0.2
	}

	words := strings.Fields(input.Text)
	if len(words) > 3 {
		wordSet := make(map[string]int)
		for _, w := range words {
			wordSet[strings.ToLower(w)]++
		}
		for _, count := range wordSet {
			if count >= 3 {
				scores[entities.EmotionConfused] += 0.4
				break
			}
		}
	}

	return scores
}

func classifyIntensity(score float64) entities.IntensityLevel {
	switch {
	case score >= entities.ThresholdHigh:
		return entities.IntensityCritical
	case score >= entities.ThresholdMedium:
		return entities.IntensityHigh
	case score >= entities.ThresholdLow:
		return entities.IntensityMedium
	default:
		return entities.IntensityLow
	}
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen]
}
