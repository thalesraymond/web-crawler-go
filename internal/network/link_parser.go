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

		token := tokenizer.Token()

		if token.Data != "a" {
			continue
		}

		for _, attr := range token.Attr {
			if attr.Key != "href" {
				continue
			}

			if !isValidLink(attr.Val) {
				continue
			}

			link, err := parsedBaseURL.Parse(attr.Val)

			if err != nil || (link.Scheme != "http" && link.Scheme != "https") {
				continue
			}

			link.Fragment = ""

			links = append(links, link.String())
		}
	}

	return links, nil
}

func isValidLink(link string) bool {
	return !(strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "mailto:") || strings.HasPrefix(link, "#"))
}
