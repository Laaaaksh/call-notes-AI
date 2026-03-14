package utils

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var phonePattern = regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// IsValidPhone checks if a string is a valid phone number (E.164 format)
func IsValidPhone(s string) bool {
	cleaned := strings.ReplaceAll(s, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	return phonePattern.MatchString(cleaned)
}

// ParseDate parses a date string in YYYY-MM-DD format
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// ParseDateTime parses a datetime string in RFC3339 format
func ParseDateTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// SanitizeString trims whitespace and limits length
func SanitizeString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
