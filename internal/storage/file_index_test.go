package storage

import (
	"os"
	"sync"
	"testing"

	"github.com/thalesraymond/web-crawler-go/internal"
	"github.com/thalesraymond/web-crawler-go/internal/indexer"
)

func makeResult(url string, words map[string]int) *internal.CrawlResult {
	tokens := make([]indexer.PageToken, 0, len(words))
	for w, c := range words {
		tokens = append(tokens, indexer.PageToken{Word: w, Count: c})
	}
	return &internal.CrawlResult{URL: url, Tokens: tokens}
}

// ---------------------------------------------------------------------------
// Add
// ---------------------------------------------------------------------------

func TestFileIndex_Add_MultiplePages(t *testing.T) {
	fi := newFileIndex("")

	if err := fi.Add(makeResult("https://a.com", map[string]int{"golang": 3, "channel": 1})); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := fi.Add(makeResult("https://b.com", map[string]int{"golang": 7})); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	postings := fi.entries["golang"]
	if len(postings) != 2 {
		t.Fatalf("expected 2 postings for 'golang', got %d", len(postings))
	}
}

func TestFileIndex_Add_NoDuplicateURL(t *testing.T) {
	fi := newFileIndex("")

	fi.Add(makeResult("https://a.com", map[string]int{"golang": 3}))
	fi.Add(makeResult("https://a.com", map[string]int{"golang": 9})) // re-crawl same page

	postings := fi.entries["golang"]
	if len(postings) != 1 {
		t.Fatalf("expected 1 posting (no duplicate), got %d", len(postings))
	}
	if postings[0].Count != 9 {
		t.Errorf("expected updated count 9, got %d", postings[0].Count)
	}
}

func TestFileIndex_Add_NilResult(t *testing.T) {
	fi := newFileIndex("")
	if err := fi.Add(nil); err == nil {
		t.Error("expected error for nil result")
	}
}

func TestFileIndex_Add_ResultWithError(t *testing.T) {
	fi := newFileIndex("")
	result := &internal.CrawlResult{URL: "https://a.com", Error: os.ErrPermission}
	if err := fi.Add(result); err == nil {
		t.Error("expected error when CrawlResult carries an error")
	}
}

func TestFileIndex_Add_EmptyTokens(t *testing.T) {
	fi := newFileIndex("")
	result := &internal.CrawlResult{URL: "https://a.com", Tokens: nil}
	if err := fi.Add(result); err != nil {
		t.Errorf("empty tokens should not be an error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Save + LoadOrCreate (round-trip)
// ---------------------------------------------------------------------------

func TestFileIndex_RoundTrip(t *testing.T) {
	tmp, err := os.CreateTemp("", "index-*.json")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	fi := newFileIndex(tmp.Name())
	fi.Add(makeResult("https://a.com", map[string]int{"wikipedia": 5}))
	fi.Add(makeResult("https://b.com", map[string]int{"wikipedia": 2, "golang": 8}))

	if err := fi.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadOrCreate(tmp.Name())
	if err != nil {
		t.Fatalf("LoadOrCreate failed: %v", err)
	}

	if got := len(loaded.entries["wikipedia"]); got != 2 {
		t.Errorf("expected 2 postings for 'wikipedia' after round-trip, got %d", got)
	}
	if got := len(loaded.entries["golang"]); got != 1 {
		t.Errorf("expected 1 posting for 'golang' after round-trip, got %d", got)
	}
}

func TestFileIndex_LoadOrCreate_NewFile(t *testing.T) {
	fi, err := LoadOrCreate("/tmp/does-not-exist-xyz.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if fi == nil {
		t.Fatal("expected non-nil FileIndex")
	}
	if len(fi.entries) != 0 {
		t.Errorf("expected empty index, got %d entries", len(fi.entries))
	}
}

func TestFileIndex_LoadOrCreate_CorruptFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "corrupt-*.json")
	if err != nil {
		t.Fatal(err)
	}
	tmp.WriteString("NOT VALID JSON{{{{")
	tmp.Close()
	defer os.Remove(tmp.Name())

	fi, err := LoadOrCreate(tmp.Name())
	if err != nil {
		t.Fatalf("expected no error for corrupt file (fallback), got: %v", err)
	}
	if len(fi.entries) != 0 {
		t.Errorf("expected empty fallback index, got %d entries", len(fi.entries))
	}
}

// ---------------------------------------------------------------------------
// Lookup
// ---------------------------------------------------------------------------

func TestFileIndex_Lookup_ExistingWord(t *testing.T) {
	fi := newFileIndex("")
	fi.Add(makeResult("https://a.com", map[string]int{"wikipedia": 3}))

	entries, err := fi.Lookup("wikipedia")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestFileIndex_Lookup_MissingWord(t *testing.T) {
	fi := newFileIndex("")

	entries, err := fi.Lookup("nonexistentword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(entries))
	}
}

func TestFileIndex_Lookup_StopWord(t *testing.T) {
	fi := newFileIndex("")

	_, err := fi.Lookup("the")
	if err == nil {
		t.Error("expected error for stop word lookup")
	}
}

func TestFileIndex_Lookup_ReturnsCopy(t *testing.T) {
	fi := newFileIndex("")
	fi.Add(makeResult("https://a.com", map[string]int{"golang": 5}))

	entries, _ := fi.Lookup("golang")
	entries[0].Count = 999 // mutate returned slice

	// internal state must not be affected
	if fi.entries["golang"][0].Count == 999 {
		t.Error("Lookup should return a copy, not a reference to internal state")
	}
}

// ---------------------------------------------------------------------------
// Thread safety
// ---------------------------------------------------------------------------

func TestFileIndex_Add_Concurrent(t *testing.T) {
	fi := newFileIndex("")
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			url := "https://page.com/" + string(rune('a'+n%26))
			fi.Add(makeResult(url, map[string]int{"concurrent": 1}))
		}(i)
	}

	wg.Wait()

	if got := len(fi.entries["concurrent"]); got == 0 {
		t.Error("expected postings after concurrent adds")
	}
}
