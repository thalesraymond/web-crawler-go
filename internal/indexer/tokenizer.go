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
	return tagName == "script" || tagName == "style" || tagName == "nav" || tagName == "footer" || tagName == "header" || tagName == "aside"
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
			name, _ := tokenizer.TagName()
			if isInvalidTag(string(name)) {
				insideInvalidTags++
			}
		}

		if tokenType == html.EndTagToken {
			name, _ := tokenizer.TagName()
			if isInvalidTag(string(name)) {
				insideInvalidTags--
			}
		}

		if tokenType == html.TextToken && insideInvalidTags == 0 {
			textBuilder.Write(tokenizer.Text())
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
		finalTokens = append(finalTokens, PageToken{Word: word, Count: count})
	}

	return finalTokens
}
