package extraction

import (
	"context"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/extraction/entities"
)

type ICore interface {
	ProcessSegment(ctx context.Context, segment *entities.TranscriptSegment) (*entities.ExtractionResult, error)
}

type Core struct {
	ruleEngine  IRuleEngine
	llmReasoner ILLMReasoner
	piiRedactor IPIIRedactor
}

var _ ICore = (*Core)(nil)

func NewCore(_ context.Context, ruleEngine IRuleEngine, llmReasoner ILLMReasoner, piiRedactor IPIIRedactor) ICore {
	return &Core{ruleEngine: ruleEngine, llmReasoner: llmReasoner, piiRedactor: piiRedactor}
}

func (c *Core) ProcessSegment(ctx context.Context, segment *entities.TranscriptSegment) (*entities.ExtractionResult, error) {
	if !segment.IsFinal {
		return &entities.ExtractionResult{SessionID: segment.SessionID}, nil
	}

	redactedText, piiMatches := c.piiRedactor.Redact(segment)
	if len(piiMatches) > 0 {
		logger.Ctx(ctx).Infow(constants.LogMsgPIIRedacted,
			constants.LogFieldSessionID, segment.SessionID,
			"pii_types_found", len(piiMatches),
		)
		segment.Text = redactedText
	}

	// Layer 1: Rule engine extraction
	ruleEntities := c.ruleEngine.Extract(segment)

	// Layer 2: Medical NER (placeholder — integrates with spaCy service)
	// nerEntities := c.medicalNER.Extract(segment)

	allEntities := make([]entities.MedicalEntity, 0, len(ruleEntities))
	allEntities = append(allEntities, ruleEntities...)

	// Layer 3: LLM reasoning — only if needed
	if c.llmReasoner.ShouldInvoke(segment, allEntities) {
		logger.Ctx(ctx).Infow(constants.LogMsgLLMInvoked,
			constants.LogFieldSessionID, segment.SessionID,
			constants.LogFieldTrigger, constants.TriggerAmbiguityOrCorrection,
		)
		llmEntities, err := c.llmReasoner.Reason(ctx, segment, allEntities)
		if err != nil {
			logger.Ctx(ctx).Warnw(constants.LogMsgLLMReasonFailed,
				constants.LogKeyError, err,
				constants.LogFieldSessionID, segment.SessionID,
			)
		} else {
			allEntities = mergeEntities(allEntities, llmEntities)
		}
	}

	return &entities.ExtractionResult{
		SessionID: segment.SessionID,
		Entities:  allEntities,
	}, nil
}

func mergeEntities(existing, newEntities []entities.MedicalEntity) []entities.MedicalEntity {
	typeMap := make(map[entities.EntityType]entities.MedicalEntity)
	for _, e := range existing {
		typeMap[e.Type] = e
	}

	for _, e := range newEntities {
		if prev, ok := typeMap[e.Type]; ok {
			if e.Confidence > prev.Confidence {
				typeMap[e.Type] = e
			}
		} else {
			typeMap[e.Type] = e
		}
	}

	merged := make([]entities.MedicalEntity, 0, len(typeMap))
	for _, e := range typeMap {
		merged = append(merged, e)
	}
	return merged
}
