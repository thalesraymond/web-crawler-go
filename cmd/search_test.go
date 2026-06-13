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

	// Create a dummy file index structure at ./data/index.json
	os.MkdirAll("./data", 0755)
	dummyIndexContent := `{"testword": [{"url_string": "http://example.com", "count": 5}]}`
	os.WriteFile("./data/index.json", []byte(dummyIndexContent), 0644)
	// Clean it up after the test runs
	defer os.Remove("./data/index.json")

	cmd := exec.Command(os.Args[0], "-test.run=TestRunSearch_Success")
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
