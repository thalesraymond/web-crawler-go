package indexer

import (
	"strings"

	"golang.org/x/net/html"
)

type PageToken struct {
	Word  string
	Count int
}

func isInvalidTag(tagName string) bool {
	lowerTag := strings.ToLower(tagName)
	return lowerTag == "script" || lowerTag == "style" || lowerTag == "nav" || lowerTag == "footer" || lowerTag == "header" || lowerTag == "aside"
}

func ExtractPageTokens(htmlBody string) []PageToken {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlBody))
	var textBuilder strings.Builder
	insideInvalidTags := 0

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			break
		}

		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if isInvalidTag(token.Data) {
				insideInvalidTags++
			}
		}

		if tokenType == html.EndTagToken {
			token := tokenizer.Token()
			if isInvalidTag(token.Data) {
				insideInvalidTags--
			}
		}

		if tokenType == html.TextToken && insideInvalidTags == 0 {
			textBuilder.Write(tokenizer.Text())
			textBuilder.WriteByte(' ')
		}
	}

	rawWords := strings.Fields(textBuilder.String())
	finalTokens := make([]PageToken, 0)

	savedWords := make(map[string]int)

	for _, word := range rawWords {
		wordToSave, err := ProcessWord(word)

		if err != nil {
			continue
		}

		savedWords[wordToSave]++
	}

	for word, count := range savedWords {
		finalTokens = append(finalTokens, PageToken{Word: word, Count: count})
	}

	return finalTokens
}
