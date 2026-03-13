package extraction

import (
	"regexp"
	"strings"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/modules/extraction/entities"
)

type IRuleEngine interface {
	Extract(segment *entities.TranscriptSegment) []entities.MedicalEntity
}

type RuleEngine struct {
	phoneRegex    *regexp.Regexp
	ageRegex      *regexp.Regexp
	durationRegex *regexp.Regexp
	severityRegex *regexp.Regexp
	hindiMedDict  map[string]string
	negationWords map[string]bool
}

func NewRuleEngine() IRuleEngine {
	return &RuleEngine{
		phoneRegex:    regexp.MustCompile(`(?:\+91[-\s]?)?[6-9]\d{9}`),
		ageRegex:      regexp.MustCompile(`(?i)(\d{1,3})\s*(?:years?|yrs?|saal|baras)\s*(?:old|ka|ki|ke)?`),
		durationRegex: regexp.MustCompile(`(?i)(\d+)\s*(din|days?|hafte|weeks?|mahine|months?|saal|years?)`),
		severityRegex: regexp.MustCompile(`(?i)(\d{1,2})\s*(?:out of|\/)\s*10`),
		hindiMedDict:  buildHindiMedicalDictionary(),
		negationWords: buildNegationWords(),
	}
}

func (r *RuleEngine) Extract(segment *entities.TranscriptSegment) []entities.MedicalEntity {
	var results []entities.MedicalEntity
	text := segment.Text

	results = append(results, r.extractPhone(text, segment)...)
	results = append(results, r.extractAge(text, segment)...)
	results = append(results, r.extractDuration(text, segment)...)
	results = append(results, r.extractSeverityNumeric(text, segment)...)
	results = append(results, r.extractHindiMedicalTerms(text, segment)...)

	return results
}

func (r *RuleEngine) extractPhone(text string, seg *entities.TranscriptSegment) []entities.MedicalEntity {
	matches := r.phoneRegex.FindAllString(text, -1)
	var results []entities.MedicalEntity
	for _, m := range matches {
		results = append(results, entities.MedicalEntity{
			Type: entities.EntityPhone, RawValue: m, NormalizedValue: m,
			Confidence: 0.95, SourceLayer: constants.SourceRuleEngine, TranscriptRef: seg.Text,
		})
	}
	return results
}

func (r *RuleEngine) extractAge(text string, seg *entities.TranscriptSegment) []entities.MedicalEntity {
	matches := r.ageRegex.FindStringSubmatch(text)
	if len(matches) < 2 {
		return nil
	}
	return []entities.MedicalEntity{{
		Type: entities.EntityAge, RawValue: matches[0], NormalizedValue: matches[1],
		Confidence: 0.92, SourceLayer: constants.SourceRuleEngine, TranscriptRef: seg.Text,
	}}
}

func (r *RuleEngine) extractDuration(text string, seg *entities.TranscriptSegment) []entities.MedicalEntity {
	matches := r.durationRegex.FindStringSubmatch(text)
	if len(matches) < 3 {
		return nil
	}
	unit := normalizeDurationUnit(matches[2])
	normalized := matches[1] + " " + unit
	return []entities.MedicalEntity{{
		Type: entities.EntityDuration, RawValue: matches[0], NormalizedValue: normalized,
		Confidence: 0.90, SourceLayer: constants.SourceRuleEngine, TranscriptRef: seg.Text,
	}}
}

func (r *RuleEngine) extractSeverityNumeric(text string, seg *entities.TranscriptSegment) []entities.MedicalEntity {
	matches := r.severityRegex.FindStringSubmatch(text)
	if len(matches) < 2 {
		return nil
	}
	return []entities.MedicalEntity{{
		Type: entities.EntitySeverity, RawValue: matches[0], NormalizedValue: matches[1] + "/10",
		Confidence: 0.93, SourceLayer: constants.SourceRuleEngine, TranscriptRef: seg.Text,
	}}
}

func (r *RuleEngine) extractHindiMedicalTerms(text string, seg *entities.TranscriptSegment) []entities.MedicalEntity {
	var results []entities.MedicalEntity
	lower := strings.ToLower(text)

	for hindi, english := range r.hindiMedDict {
		if strings.Contains(lower, hindi) {
			isNeg := r.isNegated(lower, hindi)
			results = append(results, entities.MedicalEntity{
				Type: entities.EntitySymptom, RawValue: hindi, NormalizedValue: english,
				Confidence: 0.85, SourceLayer: constants.SourceRuleEngine, TranscriptRef: seg.Text,
				IsNegated: isNeg,
			})
		}
	}
	return results
}

func (r *RuleEngine) isNegated(text, term string) bool {
	idx := strings.Index(text, term)
	if idx < 0 {
		return false
	}
	prefix := text[:idx]
	words := strings.Fields(prefix)
	lookback := 3
	if len(words) < lookback {
		lookback = len(words)
	}
	for i := len(words) - lookback; i < len(words); i++ {
		if r.negationWords[words[i]] {
			return true
		}
	}
	return false
}

func buildHindiMedicalDictionary() map[string]string {
	return map[string]string{
		"dard":                     "pain",
		"bukhar":                   "fever",
		"chakkar":                  "dizziness",
		"ulti":                     "vomiting",
		"dast":                     "diarrhea",
		"khoon":                    "bleeding",
		"sujan":                    "swelling",
		"khujli":                   "itching",
		"jalan":                    "burning sensation",
		"neend na aana":            "insomnia",
		"sans lene mein dikkat":    "difficulty breathing",
		"pet mein dard":            "abdominal pain",
		"pet mein aag":             "acid reflux",
		"sar dard":                 "headache",
		"kamar dard":               "back pain",
		"ghutne mein dard":         "knee pain",
		"aankhon mein dard":        "eye pain",
		"haath pair sunn":          "numbness in extremities",
		"thakan":                   "fatigue",
		"kamzori":                  "weakness",
		"bhookh na lagna":          "loss of appetite",
		"weight badhna":            "weight gain",
		"weight kam hona":          "weight loss",
		"khansi":                   "cough",
		"zukaam":                   "cold",
		"gala kharab":              "sore throat",
		"seene mein dard":          "chest pain",
		"peshab mein jalan":        "burning urination",
		"dawaai":                   "medication",
	}
}

func buildNegationWords() map[string]bool {
	return map[string]bool{
		"no": true, "not": true, "don't": true, "doesn't": true, "didn't": true,
		"never": true, "without": true, "nahi": true, "nhi": true, "na": true,
		"nahin": true, "mat": true, "bina": true,
	}
}

func normalizeDurationUnit(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "din"), strings.HasPrefix(lower, "day"):
		return "days"
	case strings.HasPrefix(lower, "hafte"), strings.HasPrefix(lower, "week"):
		return "weeks"
	case strings.HasPrefix(lower, "mahine"), strings.HasPrefix(lower, "month"):
		return "months"
	case strings.HasPrefix(lower, "saal"), strings.HasPrefix(lower, "year"):
		return "years"
	default:
		return raw
	}
}
