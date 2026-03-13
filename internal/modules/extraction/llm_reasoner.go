package extraction

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/modules/extraction/entities"
	"github.com/call-notes-ai-service/internal/services/llm"
)

type ILLMReasoner interface {
	ShouldInvoke(segment *entities.TranscriptSegment, existingEntities []entities.MedicalEntity) bool
	Reason(ctx context.Context, segment *entities.TranscriptSegment, existingEntities []entities.MedicalEntity) ([]entities.MedicalEntity, error)
}

type LLMReasoner struct {
	client llm.IClient
}

func NewLLMReasoner(client llm.IClient) ILLMReasoner {
	return &LLMReasoner{client: client}
}

var correctionIndicators = []string{
	"actually", "sorry", "i meant", "correction", "not right", "not left",
	"galti se", "matlab", "nahi nahi", "wo nahi",
}

var ambiguityIndicators = []string{
	"i think", "maybe", "not sure", "possibly", "shayad", "ho sakta",
	"lagta hai", "pata nahi",
}

func (r *LLMReasoner) ShouldInvoke(segment *entities.TranscriptSegment, existingEntities []entities.MedicalEntity) bool {
	lower := strings.ToLower(segment.Text)

	for _, indicator := range correctionIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}

	for _, indicator := range ambiguityIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}

	hasLowConfidence := false
	for _, e := range existingEntities {
		if e.Confidence < 0.70 {
			hasLowConfidence = true
			break
		}
	}

	return hasLowConfidence
}

func (r *LLMReasoner) Reason(ctx context.Context, segment *entities.TranscriptSegment, existingEntities []entities.MedicalEntity) ([]entities.MedicalEntity, error) {
	prompt := buildReasoningPrompt(segment, existingEntities)

	response, err := r.client.Complete(ctx, prompt)
	if err != nil {
		return nil, err
	}

	parsedEntities := parseReasoningResponse(response)

	var grounded []entities.MedicalEntity
	for _, e := range parsedEntities {
		if isGroundedInTranscript(e, segment.Text) {
			e.SourceLayer = constants.SourceLLM
			e.TranscriptRef = segment.Text
			grounded = append(grounded, e)
		}
	}

	return grounded, nil
}

func buildReasoningPrompt(segment *entities.TranscriptSegment, existing []entities.MedicalEntity) string {
	var sb strings.Builder
	sb.WriteString("You are a medical call note-taking assistant. Analyze this transcript segment and extract medical entities.\n\n")
	sb.WriteString("RULES:\n")
	sb.WriteString("- Only extract information explicitly stated in the transcript\n")
	sb.WriteString("- Handle corrections: if patient says 'actually' or 'sorry, I meant', update the entity\n")
	sb.WriteString("- Handle negations: 'no fever' means fever is ABSENT\n")
	sb.WriteString("- Return JSON array of entities with: type, raw_value, normalized_value, is_negated\n\n")

	sb.WriteString("TRANSCRIPT SEGMENT:\n")
	sb.WriteString("Speaker: " + segment.Speaker + "\n")
	sb.WriteString("Text: " + segment.Text + "\n\n")

	if len(existing) > 0 {
		sb.WriteString("PREVIOUSLY EXTRACTED ENTITIES (check for corrections/contradictions):\n")
		for _, e := range existing {
			sb.WriteString("- " + string(e.Type) + ": " + e.NormalizedValue + "\n")
		}
	}

	sb.WriteString("\nReturn ONLY a JSON array. No explanation.")
	return sb.String()
}

type llmEntity struct {
	Type            string `json:"type"`
	RawValue        string `json:"raw_value"`
	NormalizedValue string `json:"normalized_value"`
	IsNegated       bool   `json:"is_negated"`
}

func parseReasoningResponse(response string) []entities.MedicalEntity {
	trimmed := strings.TrimSpace(response)

	startIdx := strings.Index(trimmed, "[")
	endIdx := strings.LastIndex(trimmed, "]")
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return nil
	}
	jsonStr := trimmed[startIdx : endIdx+1]

	var parsed []llmEntity
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil
	}

	result := make([]entities.MedicalEntity, 0, len(parsed))
	for _, p := range parsed {
		if p.RawValue == "" {
			continue
		}
		entityType := mapLLMEntityType(p.Type)
		if entityType == "" {
			continue
		}
		result = append(result, entities.MedicalEntity{
			Type:            entityType,
			RawValue:        p.RawValue,
			NormalizedValue: p.NormalizedValue,
			Confidence:      0.75,
			IsNegated:       p.IsNegated,
		})
	}
	return result
}

func mapLLMEntityType(raw string) entities.EntityType {
	typeMap := map[string]entities.EntityType{
		"symptom":    entities.EntitySymptom,
		"body_part":  entities.EntityBodyPart,
		"condition":  entities.EntityCondition,
		"medication": entities.EntityMedication,
		"duration":   entities.EntityDuration,
		"severity":   entities.EntitySeverity,
		"age":        entities.EntityAge,
		"name":       entities.EntityName,
		"phone":      entities.EntityPhone,
		"gender":     entities.EntityGender,
		"allergy":    entities.EntityAllergy,
		"follow_up":  entities.EntityFollowUp,
		"icd10_code": entities.EntityICD10,
	}
	t, ok := typeMap[strings.ToLower(strings.TrimSpace(raw))]
	if !ok {
		return ""
	}
	return t
}

func isGroundedInTranscript(entity entities.MedicalEntity, transcript string) bool {
	lower := strings.ToLower(transcript)
	rawLower := strings.ToLower(entity.RawValue)

	if strings.Contains(lower, rawLower) {
		return true
	}

	words := strings.Fields(rawLower)
	matchCount := 0
	for _, w := range words {
		if strings.Contains(lower, w) {
			matchCount++
		}
	}

	return len(words) > 0 && float64(matchCount)/float64(len(words)) >= 0.5
}
