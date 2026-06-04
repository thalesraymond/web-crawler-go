package network

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CrawlerClient struct {
	userAgent  string
	timeout    time.Duration
	httpClient *http.Client
}

func NewCrawlerClient() *CrawlerClient {
	timeout := 10 * time.Second
	return &CrawlerClient{
		userAgent: "raymond-go-crawler/1.0",
		timeout:   timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *CrawlerClient) FetchHTML(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	httpResponse, err := c.httpClient.Do(req)

	if err != nil {
		return "", fmt.Errorf("failed to execute http request: %w", err)
	}

	defer httpResponse.Body.Close() //nolint:errcheck

	if httpResponse.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response: %d", httpResponse.StatusCode)
	}

	if !strings.HasPrefix(httpResponse.Header.Get("Content-Type"), "text/html") {
		return "", fmt.Errorf("unexpected content type: %s", httpResponse.Header.Get("Content-Type"))
	}

	limitedStream := io.LimitReader(httpResponse.Body, 10*1024*1024) // Limit to 10MB to prevent memory issues using limit decorator pattern
	bodyBytes, err := io.ReadAll(limitedStream)

	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}
