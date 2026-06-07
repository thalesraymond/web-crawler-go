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

func ProcessWord(word string) (string, error) {

	word = strings.TrimSpace(word)

	if word == "" {
		return "", fmt.Errorf("word cannot be empty")
	}

	parts := strings.Fields(word)
	if len(parts) > 1 {
		return "", fmt.Errorf("word contains more than one word")
	}

	cleanWord := strings.ToLower(word)

	cleanWord = strings.TrimRightFunc(cleanWord, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSymbol(r)
	})

	wordWithoutAccents, _, err := transform.String(normalizer, cleanWord)
	if err != nil {
		wordWithoutAccents = cleanWord
	}
	if wordWithoutAccents == "" {
		return "", nil
	}

	return wordWithoutAccents, nil
}
