package network

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func ExtractLinks(baseURL string, htmlBody string) ([]string, error) {
	var links []string

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

			link, err := parsedBaseURL.Parse(attr.Val)

			if err != nil || (link.Scheme != "http" && link.Scheme != "https") {
				continue
			}

			absURL := parsedBaseURL.ResolveReference(link)
			absURL.Fragment = ""

			links = append(links, absURL.String())
		}
	}

	return links, nil
}
