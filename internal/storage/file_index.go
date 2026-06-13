package storage

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"

	"github.com/thalesraymond/web-crawler-go/internal"
	"github.com/thalesraymond/web-crawler-go/internal/indexer"
)

// FileIndex is a JSON-file-backed inverted index.
// It satisfies both the IndexWriter and IndexReader interfaces
// defined at their respective consumer sites.
type FileIndex struct {
	entries  map[string][]internal.IndexEntry
	filePath string
	mu       sync.Mutex
}

// LoadOrCreate attempts to load an existing index from filePath.
// If the file does not exist it returns a fresh, empty index.
// If the file is corrupted it logs a warning and returns a fresh index.
func LoadOrCreate(filePath string) (*FileIndex, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return newFileIndex(filePath), nil
		}
		return nil, err
	}

	var entries map[string][]internal.IndexEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		log.Printf("warning: index file %q is corrupted (%v); starting fresh", filePath, err)
		return newFileIndex(filePath), nil
	}

	return &FileIndex{
		entries:  entries,
		filePath: filePath,
	}, nil
}

func newFileIndex(filePath string) *FileIndex {
	return &FileIndex{
		entries:  make(map[string][]internal.IndexEntry),
		filePath: filePath,
	}
}

// Add merges the tokens from a CrawlResult into the in-memory index.
// If a URL was already indexed for a given word (from a previous session),
// its count is updated rather than duplicated.
func (fi *FileIndex) Add(result *internal.CrawlResult) error {
	if result == nil {
		return errors.New("crawl result cannot be nil")
	}
	if result.Error != nil {
		return result.Error
	}
	if len(result.Tokens) == 0 {
		return nil // nothing to index — not an error
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	for _, token := range result.Tokens {
		postings := fi.entries[token.Word]
		updated := false

		for idx := range postings {
			if postings[idx].UrlString == result.URL {
				postings[idx].Count = token.Count // update in-place
				updated = true
				break
			}
		}

		if !updated {
			fi.entries[token.Word] = append(postings, internal.IndexEntry{
				UrlString: result.URL,
				Count:     token.Count,
			})
		} else {
			fi.entries[token.Word] = postings
		}
	}

	return nil
}

// Save serializes the index to disk as JSON.
func (fi *FileIndex) Save() error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	data, err := json.MarshalIndent(fi.entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fi.filePath, data, 0644)
}

// Lookup returns the posting list for the given word.
// The word is normalized through ProcessWord before the lookup,
// so callers can pass raw user input.
func (fi *FileIndex) Lookup(word string) ([]internal.IndexEntry, error) {
	normalized, err := indexer.ProcessWord(word)
	if err != nil {
		return nil, err
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	entries, ok := fi.entries[normalized]
	if !ok {
		return []internal.IndexEntry{}, nil
	}

	// return a copy to prevent the caller from mutating internal state
	result := make([]internal.IndexEntry, len(entries))
	copy(result, entries)
	return result, nil
}

// GetRandomIndexedURL returns a randomly selected URL from the existing index.
// If the index is empty, it returns an empty string and false.
func (fi *FileIndex) GetRandomIndexedURL() (string, bool) {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	for _, entries := range fi.entries {
		if len(entries) > 0 {
			// Map iteration order is pseudo-random in Go.
			// Returning the first URL we encounter is sufficient.
			return entries[0].UrlString, true
		}
	}
	return "", false
}
