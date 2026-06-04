package network

import (
	"fmt"
	"io"
	"net/http"
	"context"
	"strings"
)

func FetchHTML(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	httpClient, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	httpResponse, err := http.DefaultClient.Do(httpClient)

	defer httpResponse.Body.Close() // Ensure the response body is closed after function returns

	if httpResponse.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response: %d", httpResponse.StatusCode)
	}

	if !strings.HasPrefix(httpResponse.Header.Get("Content-Type"), "text/html") {
		return "", fmt.Errorf("unexpected content type: %s", httpResponse.Header.Get("Content-Type"))
	}

	inputStream := httpResponse.Body
	limitedStream := io.LimitReader(inputStream, 10*1024*1024) // Limit to 10MB to prevent memory issues using limit decorator pattern
	bodyBytes, err := io.ReadAll(limitedStream) 

	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}
