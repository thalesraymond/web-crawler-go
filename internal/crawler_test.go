package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/thalesraymond/web-crawler-go/internal/network"
)

// newTestServer creates an httptest.Server whose handler is provided by the caller.
// The server is automatically closed when the test ends.
func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// htmlPage returns a minimal HTML page that links to the given URLs.
func htmlPage(links ...string) string {
	body := "<html><body>"
	for _, l := range links {
		body += fmt.Sprintf(`<a href="%s">link</a>`, l)
	}
	body += "</body></html>"
	return body
}

// newCrawler builds a Crawler backed by a real CrawlerClient and a fresh URLTracker.
// crawlDelay is zeroed out so tests don't pay the production rate-limit wait.
func newCrawler(concurrency, pageLimit int) *Crawler {
	client := network.NewCrawlerClient()
	tracker := network.NewURLTracker()
	c := NewCrawler(client, tracker, concurrency, pageLimit)
	c.crawlDelay = 0
	return c
}

// TestNewCrawler verifies that NewCrawler initialises the struct correctly.
func TestNewCrawler(t *testing.T) {
	client := network.NewCrawlerClient()
	tracker := network.NewURLTracker()
	c := NewCrawler(client, tracker, 3, 10)

	if c == nil {
		t.Fatal("NewCrawler returned nil")
	}
	if c.concurrency != 3 {
		t.Errorf("concurrency: got %d, want 3", c.concurrency)
	}
	if c.pageLimit != 10 {
		t.Errorf("pageLimit: got %d, want 10", c.pageLimit)
	}
	if c.client != client {
		t.Error("client field not set correctly")
	}
	if c.urlTracker != tracker {
		t.Error("urlTracker field not set correctly")
	}
}

// TestGetResults_InitiallyEmpty verifies that a fresh Crawler has no results.
func TestGetResults_InitiallyEmpty(t *testing.T) {
	c := newCrawler(1, 5)
	if got := c.GetResults(); len(got) != 0 {
		t.Errorf("expected empty results, got %d", len(got))
	}
}

// TestStart_SinglePage crawls exactly one page and expects one result.
func TestStart_SinglePage(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlPage()) // no outbound links
	})

	c := newCrawler(1, 5)
	c.Start(srv.URL)

	results := c.GetResults()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].URL != srv.URL {
		t.Errorf("URL: got %q, want %q", results[0].URL, srv.URL)
	}
	if results[0].Error != nil {
		t.Errorf("unexpected error: %v", results[0].Error)
	}
}

// TestStart_PageLimit verifies that the crawler stops after reaching the page limit.
func TestStart_PageLimit(t *testing.T) {
	const limit = 3

	// Each page links back to itself plus a unique child page.  Using a counter
	// lets us generate an unbounded number of unique URLs so we can confirm the
	// limit is respected.
	var counter atomic.Int32

	var srv *httptest.Server
	srv = newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		id := int(counter.Add(1))
		child := fmt.Sprintf("%s/page/%d", srv.URL, id)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlPage(child))
	})

	c := newCrawler(2, limit)
	c.Start(srv.URL)

	results := c.GetResults()
	if len(results) > limit {
		t.Errorf("page limit not respected: got %d results, limit was %d", len(results), limit)
	}
}

// TestStart_NoDuplicates verifies that the same URL is not crawled twice.
func TestStart_NoDuplicates(t *testing.T) {
	var requestCount atomic.Int32

	var srv *httptest.Server
	srv = newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		// Every page links to the same single child URL, which should only be visited once.
		child := srv.URL + "/child"
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlPage(child, child, child)) // duplicated intentionally
	})

	c := newCrawler(2, 10)
	c.Start(srv.URL)

	// We expect exactly 2 requests: the seed URL and /child.
	got := int(requestCount.Load())
	if got != 2 {
		t.Errorf("expected 2 HTTP requests (no duplicates), got %d", got)
	}
}

// TestStart_HTTPError verifies that a page that returns a non-200 status is not
// added to results but does not crash the crawler.
func TestStart_HTTPError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := newCrawler(1, 5)
	c.Start(srv.URL)

	results := c.GetResults()
	if len(results) != 0 {
		t.Errorf("expected 0 results for error page, got %d", len(results))
	}
}

// TestStart_LinksExtracted verifies that links found on a page are followed.
func TestStart_LinksExtracted(t *testing.T) {
	var srv *httptest.Server
	srv = newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		switch r.URL.Path {
		case "/":
			// Root page links to two children.
			fmt.Fprint(w, htmlPage(srv.URL+"/a", srv.URL+"/b"))
		default:
			// Child pages have no outbound links.
			fmt.Fprint(w, htmlPage())
		}
	})

	c := newCrawler(2, 10)
	c.Start(srv.URL)

	results := c.GetResults()
	if len(results) != 3 {
		t.Errorf("expected 3 results (root + 2 children), got %d", len(results))
	}
}

// TestStart_ConcurrentWorkers checks that multiple workers are used by verifying
// that crawling completes successfully with concurrency > 1.
func TestStart_ConcurrentWorkers(t *testing.T) {
	const numPages = 5

	var srv *httptest.Server
	srv = newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		// Generate unique child pages so the crawler has work for all workers.
		var links []string
		for i := 0; i < numPages; i++ {
			links = append(links, fmt.Sprintf("%s/page/%d", srv.URL, i))
		}
		if r.URL.Path == "/" {
			fmt.Fprint(w, htmlPage(links...))
		} else {
			fmt.Fprint(w, htmlPage())
		}
	})

	c := newCrawler(4, numPages+1)
	c.Start(srv.URL)

	results := c.GetResults()
	// We expect the root page plus up to numPages children.
	if len(results) == 0 {
		t.Error("expected results with concurrent workers, got none")
	}
}

// TestStart_ResultContainsTokens verifies that CrawlResult.Tokens is populated.
func TestStart_ResultContainsTokens(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><p>hello world</p></body></html>`)
	})

	c := newCrawler(1, 5)
	c.Start(srv.URL)

	results := c.GetResults()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if len(results[0].Tokens) == 0 {
		t.Error("expected Tokens to be populated, got none")
	}
}

// TestStart_ResultContainsHTML verifies that the raw HTML is stored in the result.
func TestStart_ResultContainsHTML(t *testing.T) {
	const pageHTML = `<html><body>test content</body></html>`

	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, pageHTML)
	})

	c := newCrawler(1, 5)
	c.Start(srv.URL)

	results := c.GetResults()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].HTML != pageHTML {
		t.Errorf("HTML: got %q, want %q", results[0].HTML, pageHTML)
	}
}
