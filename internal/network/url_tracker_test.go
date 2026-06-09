package network

import (
	"fmt"
	"sync"
	"testing"
)

func TestURLTracker_MarkVisited(t *testing.T) {
	tracker := NewURLTracker()

	// Marking a new URL should return true
	if !tracker.MarkVisited("http://example.com") {
		t.Errorf("Expected MarkVisited to return true for a new URL")
	}

	// Marking the same URL again should return false
	if tracker.MarkVisited("http://example.com") {
		t.Errorf("Expected MarkVisited to return false for an already visited URL")
	}

	// Marking another new URL should return true
	if !tracker.MarkVisited("http://example.org") {
		t.Errorf("Expected MarkVisited to return true for a new URL")
	}
}

func TestURLTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewURLTracker()
	const numGoroutines = 100
	const numUrlsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numUrlsPerGoroutine; j++ {
				// Each goroutine will mark some unique URLs and some shared URLs
				tracker.MarkVisited(fmt.Sprintf("http://example.com/unique/%d/%d", id, j))
				tracker.MarkVisited(fmt.Sprintf("http://example.com/shared/%d", j))
			}
		}(i)
	}

	wg.Wait()

	// 100 unique URLs per goroutine (100 goroutines) = 10000 unique URLs
	// 100 shared URLs across all goroutines = 100 shared URLs
	// Total should be 10100 visited URLs
	expectedCount := numGoroutines*numUrlsPerGoroutine + numUrlsPerGoroutine
	if tracker.VisitedCount() != expectedCount {
		t.Errorf("Expected %d visited URLs, got %d", expectedCount, tracker.VisitedCount())
	}
}
