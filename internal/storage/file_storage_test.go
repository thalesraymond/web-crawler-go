package storage

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/thalesraymond/web-crawler-go/internal"
)

func newTempStorage(t *testing.T) (*FileStorage, string) {
	t.Helper()
	dir := t.TempDir()
	return NewFileStorage(dir), dir
}

func TestSave_NilResult_ReturnsError(t *testing.T) {
	s, _ := newTempStorage(t)

	err := s.Save(nil)

	if err == nil {
		t.Fatal("expected error for nil result, got nil")
	}
}

func TestSave_ValidResult_CreatesFile(t *testing.T) {
	s, dir := newTempStorage(t)

	result := &internal.CrawlResult{
		URL:   "https://example.com",
		Links: []string{"https://example.com/about"},
	}

	if err := s.Save(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	encodedURL := base64.URLEncoding.EncodeToString([]byte(result.URL))
	expectedPath := filepath.Join(dir, encodedURL+".json")

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist, but it does not", expectedPath)
	}
}

func TestSave_ValidResult_FileContainsCorrectJSON(t *testing.T) {
	s, dir := newTempStorage(t)

	result := &internal.CrawlResult{
		URL:   "https://example.com/page",
		Links: []string{"https://example.com/a", "https://example.com/b"},
	}

	if err := s.Save(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	encodedURL := base64.URLEncoding.EncodeToString([]byte(result.URL))
	filePath := filepath.Join(dir, encodedURL+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("could not read saved file: %v", err)
	}

	var saved internal.CrawlResult
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}

	if saved.URL != result.URL {
		t.Errorf("URL mismatch: got %q, want %q", saved.URL, result.URL)
	}
	if len(saved.Links) != len(result.Links) {
		t.Errorf("Links length mismatch: got %d, want %d", len(saved.Links), len(result.Links))
	}
}

func TestSave_FileNameIsBase64EncodedURL(t *testing.T) {
	s, dir := newTempStorage(t)

	// Use a URL whose base64 encoding does not contain '/' to avoid OS treating
	// the encoded filename as a nested path.
	url := "https://example.com"
	result := &internal.CrawlResult{URL: url}

	if err := s.Save(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedName := base64.URLEncoding.EncodeToString([]byte(url)) + ".json"
	expectedPath := filepath.Join(dir, expectedName)

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file %s does not exist", expectedPath)
	}
}

func TestSave_DifferentURLs_CreateSeparateFiles(t *testing.T) {
	s, dir := newTempStorage(t)

	urls := []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/page3",
	}

	for _, url := range urls {
		if err := s.Save(&internal.CrawlResult{URL: url}); err != nil {
			t.Fatalf("unexpected error saving %q: %v", url, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("could not read directory: %v", err)
	}

	if len(entries) != len(urls) {
		t.Errorf("expected %d files, got %d", len(urls), len(entries))
	}
}

func TestSave_InvalidDirectory_ReturnsError(t *testing.T) {
	s := NewFileStorage("/nonexistent/path/that/does/not/exist")

	err := s.Save(&internal.CrawlResult{URL: "https://example.com"})

	if err == nil {
		t.Fatal("expected error for invalid directory, got nil")
	}
}

func TestSave_ConcurrentSaves_NoDataRace(t *testing.T) {
	s, _ := newTempStorage(t)

	var wg sync.WaitGroup
	urls := []string{
		"https://example.com/c1",
		"https://example.com/c2",
		"https://example.com/c3",
		"https://example.com/c4",
		"https://example.com/c5",
	}

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			if err := s.Save(&internal.CrawlResult{URL: u}); err != nil {
				t.Errorf("concurrent save error for %q: %v", u, err)
			}
		}(url)
	}

	wg.Wait()
}

func TestSave_FilePermissions_Are600(t *testing.T) {
	s, dir := newTempStorage(t)

	result := &internal.CrawlResult{URL: "https://example.com/perms"}
	if err := s.Save(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	encodedURL := base64.URLEncoding.EncodeToString([]byte(result.URL))
	filePath := filepath.Join(dir, encodedURL+".json")

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("could not stat file: %v", err)
	}

	const perm600 = os.FileMode(0600)
	if info.Mode().Perm() != perm600 {
		t.Errorf("expected permissions %o, got %o", perm600, info.Mode().Perm())
	}
}
