package network

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewCrawlerClient(t *testing.T) {
	tests := []struct {
		name              string
		expectedUserAgent string
		expectedTimeout   time.Duration
	}{
		{
			name:              "default configuration",
			expectedUserAgent: "raymond-go-crawler/1.0",
			expectedTimeout:   10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client1 := NewCrawlerClient()
			if client1 == nil {
				t.Fatal("expected non-nil CrawlerClient")
			}
			if client1.userAgent != tt.expectedUserAgent {
				t.Errorf("expected userAgent %q, got %q", tt.expectedUserAgent, client1.userAgent)
			}
			if client1.timeout != tt.expectedTimeout {
				t.Errorf("expected timeout %v, got %v", tt.expectedTimeout, client1.timeout)
			}
			if client1.httpClient == nil {
				t.Fatal("expected non-nil httpClient")
			}
			if client1.httpClient.Timeout != tt.expectedTimeout {
				t.Errorf("expected httpClient.Timeout %v, got %v", tt.expectedTimeout, client1.httpClient.Timeout)
			}

			// Edge case: ensure multiple client instances do not share the same pointers
			client2 := NewCrawlerClient()
			if client1 == client2 {
				t.Errorf("expected unique CrawlerClient instances, got same pointer")
			}
			if client1.httpClient == client2.httpClient {
				t.Errorf("expected unique http.Client instances, got same pointer")
			}
		})
	}
}

func TestFetchHTML(t *testing.T) {
	client := NewCrawlerClient()

	t.Run("empty URL", func(t *testing.T) {
		_, err := client.FetchHTML(context.Background(), "")
		if err == nil {
			t.Errorf("expected error for empty URL, got nil")
		}
	})

	t.Run("successful fetch", func(t *testing.T) {
		expectedBody := "<html><body>Hello World</body></html>"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("User-Agent") != client.userAgent {
				t.Errorf("expected User-Agent %q, got %q", client.userAgent, r.Header.Get("User-Agent"))
			}
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, expectedBody)
		}))
		defer server.Close()

		body, err := client.FetchHTML(context.Background(), server.URL)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if body != expectedBody {
			t.Errorf("expected body %q, got %q", expectedBody, body)
		}
	})

	t.Run("non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		_, err := client.FetchHTML(context.Background(), server.URL)
		if err == nil {
			t.Errorf("expected error for 404 response, got nil")
		}

		if !strings.Contains(err.Error(), "404") {
			t.Errorf("expected 404 error, got %v", err)
		}
	})

	t.Run("network error", func(t *testing.T) {
		// Using an invalid URL format to trigger a network error quickly
		_, err := client.FetchHTML(context.Background(), "http://localhost:0")
		if err == nil {
			t.Errorf("expected network error, got nil")
		}
	})

	t.Run("content-type error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, "not html")
		}))
		defer server.Close()

		_, err := client.FetchHTML(context.Background(), server.URL)
		if err == nil {
			t.Errorf("expected content-type error, got nil")
		}

		if !strings.Contains(err.Error(), "unexpected content type") {
			t.Errorf("expected content-type error, got %v", err)
		}
	})
}
