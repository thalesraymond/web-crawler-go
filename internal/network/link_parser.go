package network

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func ExtractLinks(baseURL string, htmlBody string) ([]string, error) {
	var links []string

	if baseURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}

	if htmlBody == "" {
		return nil, fmt.Errorf("HTML body cannot be empty")
	}

	parsedBaseURL, err := url.Parse(baseURL)

	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	tokenizer := html.NewTokenizer(strings.NewReader(htmlBody))

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			break
		}

		if tokenType != html.StartTagToken {
			continue
		}

		tagName, hasAttr := tokenizer.TagName()

		if len(tagName) != 1 || tagName[0] != 'a' {
			continue
		}

		for hasAttr {
			var key, val []byte
			key, val, hasAttr = tokenizer.TagAttr()

			if string(key) != "href" {
				continue
			}

			attrVal := string(val)
			if !isValidLink(attrVal) {
				continue
			}

			link, err := parsedBaseURL.Parse(attrVal)

			if err != nil || (link.Scheme != "http" && link.Scheme != "https") {
				continue
			}

			link.Fragment = ""

			links = append(links, link.String())
		}
	}

	return links, nil
}

var invalidPrefixes = []string{
	"javascript:",
	"mailto:",
	"#",
}

func isValidLink(link string) bool {
	lowerLink := strings.ToLower(link)
	for _, invalidPrefix := range invalidPrefixes {
		if strings.HasPrefix(lowerLink, invalidPrefix) {
			return false
		}
	}
	return true
}
