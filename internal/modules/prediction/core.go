package prediction

import (
	"context"
	"math"
	"time"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/prediction/entities"
)

type ICore interface {
	GetPredictedFields(ctx context.Context, phone string) (*entities.PatientHistoryResponse, error)
}

type Core struct {
	repo IRepository
}

var _ ICore = (*Core)(nil)

func NewCore(repo IRepository) ICore {
	return &Core{repo: repo}
}

func (c *Core) GetPredictedFields(ctx context.Context, phone string) (*entities.PatientHistoryResponse, error) {
	history, err := c.repo.GetPatientHistory(ctx, phone)
	if err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgPredictionLookupFailed,
			constants.LogKeyError, err,
			constants.LogFieldPatientPhone, phone,
		)
		return nil, err
	}

	sessionCount, err := c.repo.GetSessionCountByPhone(ctx, phone)
	if err != nil {
		sessionCount = 0
	}

	predicted := make([]entities.PredictedField, 0, len(history))
	var lastVisit *time.Time

	for _, entry := range history {
		confidence := calculateConfidence(entry)
		if confidence < entities.MinPreFillConfidence {
			continue
		}

		predicted = append(predicted, entities.PredictedField{
			FieldName:       entry.FieldName,
			FieldValue:      entry.FieldValue,
			Confidence:      confidence,
			Source:          entities.SourceHistory,
			OccurrenceCount: entry.OccurrenceCount,
			LastSeenAt:      entry.LastSeenAt,
		})

		if lastVisit == nil || entry.LastSeenAt.After(*lastVisit) {
			t := entry.LastSeenAt
			lastVisit = &t
		}
	}

	logger.Ctx(ctx).Infow(constants.LogMsgPredictionComplete,
		constants.LogFieldPatientPhone, phone,
		constants.LogFieldPredictedCount, len(predicted),
		constants.LogFieldSessionCount, sessionCount,
	)

	return &entities.PatientHistoryResponse{
		PatientPhone:    phone,
		TotalSessions:   sessionCount,
		PredictedFields: predicted,
		LastVisitDate:   lastVisit,
	}, nil
}

func calculateConfidence(entry entities.PatientHistoryEntry) float64 {
	monthsSinceLastSeen := time.Since(entry.LastSeenAt).Hours() / (24 * 30)

	decayFactor := math.Max(
		entities.MinDecayFactor,
		1.0-(monthsSinceLastSeen*entities.MonthlyDecayRate),
	)

	repetitionBoost := math.Min(
		entities.MaxRepetitionBoost,
		1.0+float64(entry.OccurrenceCount-1)*entities.RepetitionBoostPerOccurrence,
	)

	confidence := entities.BaseHistoryConfidence * decayFactor * repetitionBoost

	return math.Round(confidence*1000) / 1000
}
