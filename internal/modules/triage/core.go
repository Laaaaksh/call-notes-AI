package triage

import (
	"context"
	"strings"
	"time"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/triage/entities"
	"github.com/google/uuid"
)

type ICore interface {
	EvaluateSymptom(ctx context.Context, sessionID uuid.UUID, input *entities.TriageInput) (*entities.TriageAssessment, error)
	GetAssessment(ctx context.Context, sessionID uuid.UUID) (*entities.TriageAssessment, error)
}

type Core struct {
	repo         IRepository
	symptomTable map[string]symptomEntry
	redFlags     map[string]bool
}

type symptomEntry struct {
	baseScore int
	isRedFlag bool
}

var _ ICore = (*Core)(nil)

func NewCore(repo IRepository) ICore {
	c := &Core{
		repo:         repo,
		symptomTable: buildSymptomTable(),
		redFlags:     buildRedFlags(),
	}
	return c
}

func (c *Core) EvaluateSymptom(ctx context.Context, sessionID uuid.UUID, input *entities.TriageInput) (*entities.TriageAssessment, error) {
	existing, err := c.repo.GetLatestAssessment(ctx, sessionID)
	if err != nil {
		existing = &entities.TriageAssessment{
			ID:        uuid.New(),
			SessionID: sessionID,
			Version:   0,
			Symptoms:  []entities.SymptomScore{},
			RedFlags:  []string{},
		}
	}

	symptomKey := strings.ToLower(strings.TrimSpace(input.Symptom))
	entry, found := c.symptomTable[symptomKey]
	baseScore := 1
	isRedFlag := false
	if found {
		baseScore = entry.baseScore
		isRedFlag = entry.isRedFlag
	}

	if input.IsResolved {
		baseScore += entities.ModifierResolvedSymptomPenalty
		if baseScore < 0 {
			baseScore = 0
		}
	}

	newSymptom := entities.SymptomScore{
		Symptom:   input.Symptom,
		BaseScore: baseScore,
		IsRedFlag: isRedFlag,
	}

	alreadyExists := false
	for i, s := range existing.Symptoms {
		if strings.EqualFold(s.Symptom, input.Symptom) {
			existing.Symptoms[i] = newSymptom
			alreadyExists = true
			break
		}
	}
	if !alreadyExists {
		existing.Symptoms = append(existing.Symptoms, newSymptom)
	}

	if isRedFlag {
		flagExists := false
		for _, rf := range existing.RedFlags {
			if strings.EqualFold(rf, input.Symptom) {
				flagExists = true
				break
			}
		}
		if !flagExists {
			existing.RedFlags = append(existing.RedFlags, input.Symptom)
		}
	}

	totalScore := 0
	for _, s := range existing.Symptoms {
		totalScore += s.BaseScore
	}

	var modifiers []entities.ModifierEntry
	modifiers = append(modifiers, c.calculateModifiers(input)...)

	modifierTotal := 0
	for _, m := range modifiers {
		modifierTotal += m.Adjustment
	}
	totalScore += modifierTotal

	if totalScore < 0 {
		totalScore = 0
	}

	urgency := scoreToUrgency(totalScore)

	now := time.Now().UTC()
	assessment := &entities.TriageAssessment{
		ID:               existing.ID,
		SessionID:        sessionID,
		UrgencyLevel:     urgency,
		CompositeScore:   totalScore,
		Symptoms:         existing.Symptoms,
		RedFlags:         existing.RedFlags,
		ModifiersApplied: modifiers,
		Version:          existing.Version + 1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := c.repo.UpsertAssessment(ctx, assessment); err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgTriageAssessmentFailed,
			constants.LogKeyError, err,
			constants.LogFieldSessionID, sessionID.String(),
		)
		return nil, err
	}

	logger.Ctx(ctx).Infow(constants.LogMsgTriageUpdated,
		constants.LogFieldSessionID, sessionID.String(),
		constants.LogFieldUrgencyLevel, string(urgency),
		constants.LogFieldTriageScore, totalScore,
	)

	return assessment, nil
}

func (c *Core) GetAssessment(ctx context.Context, sessionID uuid.UUID) (*entities.TriageAssessment, error) {
	return c.repo.GetLatestAssessment(ctx, sessionID)
}

func (c *Core) calculateModifiers(input *entities.TriageInput) []entities.ModifierEntry {
	var mods []entities.ModifierEntry

	if input.PatientAge > entities.ElderlyAgeThreshold {
		mods = append(mods, entities.ModifierEntry{
			Modifier: "elderly_patient", Adjustment: entities.ModifierElderly,
			Reason: "Patient age > 60",
		})
	}
	if input.PatientAge > 0 && input.PatientAge < entities.PediatricAgeThreshold {
		mods = append(mods, entities.ModifierEntry{
			Modifier: "pediatric_patient", Adjustment: entities.ModifierPediatric,
			Reason: "Patient age < 5",
		})
	}
	if input.DurationDays > 0 && input.DurationDays < 1 {
		mods = append(mods, entities.ModifierEntry{
			Modifier: "acute_onset", Adjustment: entities.ModifierAcuteOnset,
			Reason: "Symptom onset < 24 hours",
		})
	}
	if input.DurationDays > 14 {
		mods = append(mods, entities.ModifierEntry{
			Modifier: "chronic_escalation", Adjustment: entities.ModifierChronicEscalation,
			Reason: "Symptoms worsening > 2 weeks",
		})
	}
	if input.SymptomCount >= entities.MultipleSymptomThreshold {
		mods = append(mods, entities.ModifierEntry{
			Modifier: "multiple_symptoms", Adjustment: entities.ModifierMultipleSymptoms,
			Reason: "3+ distinct symptoms reported",
		})
	}
	if input.SentimentIntensity > 0.7 {
		mods = append(mods, entities.ModifierEntry{
			Modifier: "high_distress", Adjustment: entities.ModifierHighDistress,
			Reason: "Patient showing high emotional distress",
		})
	}
	if input.PregnancyMentioned {
		mods = append(mods, entities.ModifierEntry{
			Modifier: "pregnancy", Adjustment: entities.ModifierPregnancy,
			Reason: "Pregnancy mentioned",
		})
	}

	return mods
}

func scoreToUrgency(score int) entities.UrgencyLevel {
	switch {
	case score >= entities.ScoreThresholdCritical:
		return entities.UrgencyCritical
	case score >= entities.ScoreThresholdHigh:
		return entities.UrgencyHigh
	case score >= entities.ScoreThresholdMedium:
		return entities.UrgencyMedium
	default:
		return entities.UrgencyLow
	}
}

func buildSymptomTable() map[string]symptomEntry {
	return map[string]symptomEntry{
		"chest pain":              {9, true},
		"chest tightness":        {9, true},
		"difficulty breathing":   {9, true},
		"dyspnea":               {9, true},
		"loss of consciousness":  {10, true},
		"suicidal ideation":      {10, true},
		"severe bleeding":        {8, true},
		"sans lene mein dikkat":  {9, true},
		"behoshi":                {10, true},
		"high fever":             {7, false},
		"severe abdominal pain":  {7, false},
		"persistent vomiting":    {6, false},
		"severe headache":        {7, false},
		"seizure":                {8, true},
		"stroke symptoms":        {10, true},
		"allergic reaction":      {8, true},
		"joint pain":             {3, false},
		"knee pain":              {3, false},
		"back pain":              {3, false},
		"mild headache":          {2, false},
		"skin rash":              {2, false},
		"common cold":            {1, false},
		"cough":                  {2, false},
		"runny nose":             {1, false},
		"sore throat":            {2, false},
		"dard":                   {3, false},
		"bukhar":                 {5, false},
		"ulti":                   {4, false},
		"chakkar":                {5, false},
	}
}

func buildRedFlags() map[string]bool {
	return map[string]bool{
		"chest pain":             true,
		"chest tightness":       true,
		"difficulty breathing":  true,
		"loss of consciousness": true,
		"suicidal ideation":     true,
		"severe bleeding":       true,
		"seizure":               true,
		"stroke symptoms":       true,
		"allergic reaction":     true,
	}
}
