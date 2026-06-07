package indexer

import (
	"strings"

	"golang.org/x/net/html"
)

type PageToken struct {
	word  string
	count int
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
			if token.Data == "script" || token.Data == "style" || token.Data == "nav" || token.Data == "footer" || token.Data == "header" || token.Data == "aside" {
				insideInvalidTags++
			}
		}

		if tokenType == html.EndTagToken {
			token := tokenizer.Token()
			if token.Data == "script" || token.Data == "style" || token.Data == "nav" || token.Data == "footer" || token.Data == "header" || token.Data == "aside" {
				insideInvalidTags--
			}
		}

		if tokenType == html.TextToken && insideInvalidTags == 0 {
			textBuilder.WriteString(tokenizer.Token().Data)
			textBuilder.WriteString(" ")
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
		finalTokens = append(finalTokens, PageToken{word: word, count: count})
	}

	return finalTokens
}
