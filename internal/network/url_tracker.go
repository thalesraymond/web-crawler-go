package network

import (
	"sync"
)

type URLTracker struct {
	mu      sync.Mutex
	visited map[string]struct{}
}

func NewURLTracker() *URLTracker {
	return &URLTracker{
		mu:      sync.Mutex{},
		visited: make(map[string]struct{}),
	}
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
