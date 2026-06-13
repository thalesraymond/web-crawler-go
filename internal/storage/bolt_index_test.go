package storage

import (
	"os"
	"sync"
	"testing"

	"github.com/thalesraymond/web-crawler-go/internal"
)

// Helper to create an empty test BoltIndex in a temporary file.
func newBoltIndexForTest(t *testing.T) (*BoltIndex, func()) {
	tmp, err := os.CreateTemp("", "bolt-index-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close() // nolint:errcheck

	bi, err := LoadOrCreateBolt(tmp.Name())
	if err != nil {
		t.Fatalf("LoadOrCreateBolt failed: %v", err)
	}

	cleanup := func() {
		bi.Close() // nolint:errcheck
		os.Remove(tmp.Name()) // nolint:errcheck
	}

	return bi, cleanup
}

// ---------------------------------------------------------------------------
// Add
// ---------------------------------------------------------------------------

func TestBoltIndex_Add_MultiplePages(t *testing.T) {
	bi, cleanup := newBoltIndexForTest(t)
	defer cleanup()

	if err := bi.Add(makeResult("https://a.com", map[string]int{"golang": 3, "channel": 1})); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := bi.Add(makeResult("https://b.com", map[string]int{"golang": 7})); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	postings, err := bi.Lookup("golang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(postings) != 2 {
		t.Fatalf("expected 2 postings for 'golang', got %d", len(postings))
	}
}

func TestBoltIndex_Add_NoDuplicateURL(t *testing.T) {
	bi, cleanup := newBoltIndexForTest(t)
	defer cleanup()

	bi.Add(makeResult("https://a.com", map[string]int{"golang": 3})) // nolint:errcheck
	bi.Add(makeResult("https://a.com", map[string]int{"golang": 9})) // nolint:errcheck

	postings, err := bi.Lookup("golang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(postings) != 1 {
		t.Fatalf("expected 1 posting (no duplicate), got %d", len(postings))
	}
	if postings[0].Count != 9 {
		t.Errorf("expected updated count 9, got %d", postings[0].Count)
	}
}

func TestBoltIndex_Add_NilResult(t *testing.T) {
	bi, cleanup := newBoltIndexForTest(t)
	defer cleanup()

	if err := bi.Add(nil); err == nil {
		t.Error("expected error for nil result")
	}
}

func TestBoltIndex_Add_ResultWithError(t *testing.T) {
	bi, cleanup := newBoltIndexForTest(t)
	defer cleanup()

	result := &internal.CrawlResult{URL: "https://a.com", Error: os.ErrPermission}
	if err := bi.Add(result); err == nil {
		t.Error("expected error when CrawlResult carries an error")
	}
}

func TestBoltIndex_Add_EmptyTokens(t *testing.T) {
	bi, cleanup := newBoltIndexForTest(t)
	defer cleanup()

	result := &internal.CrawlResult{URL: "https://a.com", Tokens: nil}
	if err := bi.Add(result); err != nil {
		t.Errorf("empty tokens should not be an error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// LoadOrCreateBolt
// ---------------------------------------------------------------------------

func TestBoltIndex_RoundTrip(t *testing.T) {
	tmp, err := os.CreateTemp("", "bolt-index-roundtrip-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close() // nolint:errcheck
	defer os.Remove(tmp.Name()) // nolint:errcheck

	bi, err := LoadOrCreateBolt(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	bi.Add(makeResult("https://a.com", map[string]int{"wikipedia": 5})) // nolint:errcheck
	bi.Add(makeResult("https://b.com", map[string]int{"wikipedia": 2, "golang": 8})) // nolint:errcheck

	if err := bi.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	bi.Close() // nolint:errcheck

	// Load again
	loaded, err := LoadOrCreateBolt(tmp.Name())
	if err != nil {
		t.Fatalf("LoadOrCreateBolt failed: %v", err)
	}
	defer func() { _ = loaded.Close() }()

	postingsWiki, _ := loaded.Lookup("wikipedia")
	if got := len(postingsWiki); got != 2 {
		t.Errorf("expected 2 postings for 'wikipedia' after round-trip, got %d", got)
	}
	postingsGo, _ := loaded.Lookup("golang")
	if got := len(postingsGo); got != 1 {
		t.Errorf("expected 1 posting for 'golang' after round-trip, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Lookup
// ---------------------------------------------------------------------------

func TestBoltIndex_Lookup_ExistingWord(t *testing.T) {
	bi, cleanup := newBoltIndexForTest(t)
	defer cleanup()

	bi.Add(makeResult("https://a.com", map[string]int{"wikipedia": 3})) // nolint:errcheck

	entries, err := bi.Lookup("wikipedia")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestBoltIndex_Lookup_MissingWord(t *testing.T) {
	bi, cleanup := newBoltIndexForTest(t)
	defer cleanup()

	entries, err := bi.Lookup("nonexistentword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(entries))
	}
}

func TestBoltIndex_Lookup_StopWord(t *testing.T) {
	bi, cleanup := newBoltIndexForTest(t)
	defer cleanup()

	_, err := bi.Lookup("the")
	if err == nil {
		t.Error("expected error for stop word lookup")
	}
}

func TestBoltIndex_Lookup_ReturnsCopy(t *testing.T) {
	bi, cleanup := newBoltIndexForTest(t)
	defer cleanup()

	bi.Add(makeResult("https://a.com", map[string]int{"golang": 5})) // nolint:errcheck

	entries, _ := bi.Lookup("golang")
	entries[0].Count = 999 // mutate returned slice

	// internal state must not be affected
	entries2, _ := bi.Lookup("golang")
	if entries2[0].Count == 999 {
		t.Error("Lookup should return a copy, not a reference to internal state")
	}
}

// ---------------------------------------------------------------------------
// Thread safety
// ---------------------------------------------------------------------------

func TestBoltIndex_Add_Concurrent(t *testing.T) {
	bi, cleanup := newBoltIndexForTest(t)
	defer cleanup()

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			url := "https://page.com/" + string(rune('a'+n%26))
			bi.Add(makeResult(url, map[string]int{"concurrent": 1})) // nolint:errcheck
		}(i)
	}

	wg.Wait()

	postings, _ := bi.Lookup("concurrent")
	if got := len(postings); got == 0 {
		t.Error("expected postings after concurrent adds")
	}
}
