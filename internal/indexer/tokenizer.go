package indexer

import (
	"strings"

	"golang.org/x/net/html"
)

var stopWords = map[string]struct{}{
	// Portuguese
	"o": {}, "a": {}, "os": {}, "as": {}, "um": {}, "uma": {},
	"de": {}, "do": {}, "da": {}, "em": {}, "no": {}, "na": {},
	"para": {}, "com": {}, "por": {}, "que": {}, "e": {}, "ou": {},
	// English
	"the": {}, "of": {}, "to": {}, "and": {}, "in": {}, "is": {},
	"it": {}, "that": {}, "for": {}, "on": {}, "are": {}, "with": {},
}

type PageToken struct {
	word  string
	count int
}

func ExtractPageTokens(htmlBody string) []PageToken {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlBody))
	var textBuilder strings.Builder
	isInsideScriptOrStyle := false

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			break
		}

		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "script" || token.Data == "style" || token.Data == "nav" || token.Data == "footer" || token.Data == "header" || token.Data == "aside" {
				isInsideScriptOrStyle = true
			}
		}

		if tokenType == html.EndTagToken {
			token := tokenizer.Token()
			if token.Data == "script" || token.Data == "style" || token.Data == "nav" || token.Data == "footer" || token.Data == "header" || token.Data == "aside" {
				isInsideScriptOrStyle = false
			}
		}

		if tokenType == html.TextToken && !isInsideScriptOrStyle {
			textBuilder.WriteString(tokenizer.Token().Data)
			textBuilder.WriteString(" ")
		}
	}

	rawWords := strings.Fields(textBuilder.String())
	finalTokens := make([]PageToken, 0)

	savedWords := make(map[string]int)

	for _, word := range rawWords {
		if len(word) < 2 {
			continue
		}

		if _, isStopWord := stopWords[word]; isStopWord {
			continue
		}

		wordToSave, err := ProcessWord(word)

		if err != nil {
			continue
		}

		savedWords[wordToSave]++
	}

	for word, count := range savedWords {
		finalTokens = append(finalTokens, PageToken{word: word, count: count})
	}

	return finalTokens
}
