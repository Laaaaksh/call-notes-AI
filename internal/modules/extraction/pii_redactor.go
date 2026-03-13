package extraction

import (
	"regexp"
	"strings"

	"github.com/call-notes-ai-service/internal/modules/extraction/entities"
)

type PIIType string

const (
	PIIAadhaar    PIIType = "AADHAAR"
	PIIPan        PIIType = "PAN"
	PIICreditCard PIIType = "CREDIT_CARD"
	PIIBankAcct   PIIType = "BANK_ACCOUNT"
	PIIEmail      PIIType = "EMAIL"
)

type PIIMatch struct {
	Type       PIIType
	StartIndex int
	EndIndex   int
}

type IPIIRedactor interface {
	Redact(segment *entities.TranscriptSegment) (redacted string, found []PIIMatch)
	ContainsPII(text string) bool
}

type PIIRedactor struct {
	patterns map[PIIType]*regexp.Regexp
}

func NewPIIRedactor() IPIIRedactor {
	return &PIIRedactor{
		patterns: map[PIIType]*regexp.Regexp{
			PIIAadhaar:    regexp.MustCompile(`\b\d{4}\s?\d{4}\s?\d{4}\b`),
			PIIPan:        regexp.MustCompile(`\b[A-Z]{5}\d{4}[A-Z]\b`),
			PIICreditCard: regexp.MustCompile(`\b(?:\d{4}[\s-]?){3}\d{4}\b`),
			PIIBankAcct:   regexp.MustCompile(`\b\d{9,18}\b`),
			PIIEmail:      regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`),
		},
	}
}

func (r *PIIRedactor) Redact(segment *entities.TranscriptSegment) (string, []PIIMatch) {
	text := segment.Text
	var allMatches []PIIMatch

	for piiType, pattern := range r.patterns {
		if piiType == PIIBankAcct {
			continue
		}
		locations := pattern.FindAllStringIndex(text, -1)
		for _, loc := range locations {
			allMatches = append(allMatches, PIIMatch{
				Type:       piiType,
				StartIndex: loc[0],
				EndIndex:   loc[1],
			})
		}
	}

	if hasBankContext(text) {
		locations := r.patterns[PIIBankAcct].FindAllStringIndex(text, -1)
		for _, loc := range locations {
			allMatches = append(allMatches, PIIMatch{
				Type:       PIIBankAcct,
				StartIndex: loc[0],
				EndIndex:   loc[1],
			})
		}
	}

	if len(allMatches) == 0 {
		return text, nil
	}

	redacted := text
	for i := len(allMatches) - 1; i >= 0; i-- {
		m := allMatches[i]
		replacement := "[REDACTED_" + string(m.Type) + "]"
		redacted = redacted[:m.StartIndex] + replacement + redacted[m.EndIndex:]
	}

	return redacted, allMatches
}

func (r *PIIRedactor) ContainsPII(text string) bool {
	for piiType, pattern := range r.patterns {
		if piiType == PIIBankAcct {
			continue
		}
		if pattern.MatchString(text) {
			return true
		}
	}

	if hasBankContext(text) && r.patterns[PIIBankAcct].MatchString(text) {
		return true
	}

	return false
}

func hasBankContext(text string) bool {
	lower := strings.ToLower(text)
	bankKeywords := []string{
		"account number", "account no", "bank account",
		"khata number", "khata no", "a/c",
	}
	for _, kw := range bankKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
