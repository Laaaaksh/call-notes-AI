package sentiment

import (
	"strings"

	"github.com/call-notes-ai-service/internal/modules/sentiment/entities"
)

type Lexicon struct {
	keywords map[entities.EmotionType][]string
}

func NewLexicon() *Lexicon {
	return &Lexicon{
		keywords: map[entities.EmotionType][]string{
			entities.EmotionDistressed: {
				"help", "please", "scared", "worried", "afraid", "panic",
				"don't know what to do", "can't take it", "unbearable",
				"madad", "darr", "ghabra", "kya karu", "sahansahi",
				"dard ho raha", "bahut dard", "bardasht nahi",
			},
			entities.EmotionAngry: {
				"unacceptable", "ridiculous", "terrible", "useless", "waste",
				"manager", "supervisor", "complaint", "sue", "report",
				"kya bakwas", "bekar", "manager se baat karao",
				"complaint karunga", "pagal", "cheat",
			},
			entities.EmotionConfused: {
				"don't understand", "what do you mean", "confused",
				"can you repeat", "what?", "huh", "sorry?",
				"samajh nahi aaya", "kya?", "phir se boliye",
				"matlab?", "kaise?", "clear nahi hai",
			},
			entities.EmotionSad: {
				"unfortunately", "difficult", "suffering", "hopeless",
				"depressed", "lonely", "tired of this", "given up",
				"mushkil", "takleef", "thak gaya", "umeed nahi",
				"bahut bura", "dukhi", "pareshan",
			},
		},
	}
}

func (l *Lexicon) Score(text string) map[entities.EmotionType]float64 {
	scores := make(map[entities.EmotionType]float64)
	lower := strings.ToLower(text)

	for emotion, keywords := range l.keywords {
		matchCount := 0
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				matchCount++
			}
		}
		if matchCount > 0 {
			score := float64(matchCount) / float64(len(keywords))
			if score > 1.0 {
				score = 1.0
			}
			scores[emotion] = score * 2.0
			if scores[emotion] > 1.0 {
				scores[emotion] = 1.0
			}
		}
	}

	return scores
}
