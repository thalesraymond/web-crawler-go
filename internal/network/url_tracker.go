package network

import (
	"sync"
)

type URLTracker struct {
	mu      sync.RWMutex
	visited map[string]struct{}
}

func NewURLTracker() *URLTracker {
	return &URLTracker{
		visited: make(map[string]struct{}),
	}
}

func (t *URLTracker) VisitedCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.visited)
}

func (t *URLTracker) MarkVisited(url string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, visited := t.visited[url]; visited {
		return false
	}

	t.visited[url] = struct{}{}
	return true
}
