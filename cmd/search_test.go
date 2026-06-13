package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestRunSearch_MissingQuery verifies that running search without a query fails correctly
func TestRunSearch_MissingQuery(t *testing.T) {
	if os.Getenv("TEST_RUN_SEARCH_MISSING_QUERY") == "1" {
		runSearch([]string{})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRunSearch_MissingQuery")
	cmd.Env = append(os.Environ(), "TEST_RUN_SEARCH_MISSING_QUERY=1")
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("expected error (exit code 1), got nil")
	}

	if !strings.Contains(string(out), "Search query is required") {
		t.Errorf("expected error message about required query, got: %s", string(out))
	}
}

// TestRunSearch_NoResults verifies the behavior when querying a word that isn't in the index
func TestRunSearch_NoResults(t *testing.T) {
	if os.Getenv("TEST_RUN_SEARCH_NO_RESULTS") == "1" {
		runSearch([]string{"-query", "thiswillnotbefound_xyz", "-index-type", "file"})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRunSearch_NoResults")
	cmd.Env = append(os.Environ(), "TEST_RUN_SEARCH_NO_RESULTS=1")
	out, err := cmd.CombinedOutput()

	// Should exit with 0 because "No results" or "No pages found" is not considered a failure exit code in runSearch
	if err != nil {
		t.Fatalf("expected no error (exit 0), got %v\nOutput: %s", err, string(out))
	}

	if !strings.Contains(string(out), "No results") && !strings.Contains(string(out), "No pages found") {
		t.Errorf("expected no results message, got: %s", string(out))
	}
}

// TestRunSearch_Success verifies that results are found and correctly formatted
func TestRunSearch_Success(t *testing.T) {
	if os.Getenv("TEST_RUN_SEARCH_SUCCESS") == "1" {
		// Run search on the word "testword"
		runSearch([]string{"-query", "testword", "-index-type", "file"})
		return
	}

	tempDir := t.TempDir()
	// Create a dummy file index structure at ./data/index.json inside the temp dir
	if err := os.MkdirAll(tempDir+"/data", 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	dummyIndexContent := `{"testword": [{"url_string": "http://example.com", "count": 5}]}`
	if err := os.WriteFile(tempDir+"/data/index.json", []byte(dummyIndexContent), 0644); err != nil {
		t.Fatalf("failed to write dummy index: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRunSearch_Success")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "TEST_RUN_SEARCH_SUCCESS=1")
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("expected no error (exit 0), got %v\nOutput: %s", err, string(out))
	}

	// Verify output formats correctly
	if !strings.Contains(string(out), "Results for \"testword\"") {
		t.Errorf("expected results message, got: %s", string(out))
	}
	if !strings.Contains(string(out), "http://example.com") {
		t.Errorf("expected url in output, got: %s", string(out))
	}
	if !strings.Contains(string(out), "(count: 5)") {
		t.Errorf("expected count in output, got: %s", string(out))
	}
}

// TestRunSearch_MultipleResults verifies sorting of results by count
func TestRunSearch_MultipleResults(t *testing.T) {
	if os.Getenv("TEST_RUN_SEARCH_MULTIPLE_RESULTS") == "1" {
		runSearch([]string{"-query", "testword", "-index-type", "file"})
		return
	}

	tempDir := t.TempDir()
	if err := os.MkdirAll(tempDir+"/data", 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	dummyIndexContent := `{"testword": [{"url_string": "http://example.com/b", "count": 2}, {"url_string": "http://example.com/a", "count": 5}, {"url_string": "http://example.com/c", "count": 1}]}`
	if err := os.WriteFile(tempDir+"/data/index.json", []byte(dummyIndexContent), 0644); err != nil {
		t.Fatalf("failed to write dummy index: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRunSearch_MultipleResults")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "TEST_RUN_SEARCH_MULTIPLE_RESULTS=1")
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("expected no error (exit 0), got %v\nOutput: %s", err, string(out))
	}

	outStr := string(out)

	idxA := strings.Index(outStr, "http://example.com/a")
	idxB := strings.Index(outStr, "http://example.com/b")
	idxC := strings.Index(outStr, "http://example.com/c")

	if idxA == -1 || idxB == -1 || idxC == -1 {
		t.Fatalf("expected all urls in output, got: %s", outStr)
	}

	if !(idxA < idxB && idxB < idxC) {
		t.Errorf("expected results to be sorted by count (a:5, b:2, c:1), but got order a:%d, b:%d, c:%d in output: %s", idxA, idxB, idxC, outStr)
	}
}

// TestRunSearch_BoltDB verifies search with boltdb index type
func TestRunSearch_BoltDB(t *testing.T) {
	if os.Getenv("TEST_RUN_SEARCH_BOLTDB") == "1" {
		runSearch([]string{"-query", "testword", "-index-type", "bolt"})
		return
	}

	tempDir := t.TempDir()
	if err := os.MkdirAll(tempDir+"/data", 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRunSearch_BoltDB")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "TEST_RUN_SEARCH_BOLTDB=1")
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("expected no error (exit 0), got %v\nOutput: %s", err, string(out))
	}

	if !strings.Contains(string(out), "No results") && !strings.Contains(string(out), "No pages found") {
		t.Errorf("expected no results message, got: %s", string(out))
	}
}

// TestRunSearch_IndexLoadError verifies that a load error is handled correctly
func TestRunSearch_IndexLoadError(t *testing.T) {
	if os.Getenv("TEST_RUN_SEARCH_INDEX_LOAD_ERROR") == "1" {
		runSearch([]string{"-query", "testword", "-index-type", "file"})
		return
	}

	tempDir := t.TempDir()
	// Create a directory instead of a file to force an error when trying to read index.json
	if err := os.MkdirAll(tempDir+"/data/index.json", 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRunSearch_IndexLoadError")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "TEST_RUN_SEARCH_INDEX_LOAD_ERROR=1")
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("expected error due to log.Fatalf, got nil\nOutput: %s", string(out))
	}

	if !strings.Contains(string(out), "Error loading index") {
		t.Errorf("expected 'Error loading index' in output, got: %s", string(out))
	}
}
