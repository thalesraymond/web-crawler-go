package network

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
