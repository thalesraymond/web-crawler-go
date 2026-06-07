package indexer

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var normalizer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

var stopWords = map[string]struct{}{
	// Portuguese
	"o": {}, "a": {}, "os": {}, "as": {}, "um": {}, "uma": {},
	"de": {}, "do": {}, "da": {}, "em": {}, "no": {}, "na": {},
	"para": {}, "com": {}, "por": {}, "que": {}, "e": {}, "ou": {},
	// English
	"the": {}, "of": {}, "to": {}, "and": {}, "in": {}, "is": {},
	"it": {}, "that": {}, "for": {}, "on": {}, "are": {}, "with": {},
}

func ProcessWord(word string) (string, error) {

	cleanWord := strings.ToLower(strings.TrimSpace(word))

	if len(cleanWord) < 2 {
		return "", fmt.Errorf("word is too short")
	}

	parts := strings.Fields(cleanWord)
	if len(parts) > 1 {
		return "", fmt.Errorf("word contains more than one word")
	}

	cleanWord = strings.TrimRightFunc(cleanWord, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSymbol(r)
	})

	wordWithoutAccents, _, err := transform.String(normalizer, cleanWord)
	if err != nil {
		wordWithoutAccents = cleanWord
	}
	if wordWithoutAccents == "" {
		return "", fmt.Errorf("word is empty after normalization")
	}

	if _, isStopWord := stopWords[wordWithoutAccents]; isStopWord {
		return "", fmt.Errorf("word is a stop word")
	}

	return wordWithoutAccents, nil
}
