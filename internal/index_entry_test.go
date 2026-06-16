package internal

import (
	"encoding/json"
	"testing"
)

func TestIndexEntry_JSON(t *testing.T) {
	tests := []struct {
		name     string
		entry    IndexEntry
		expected string
	}{
		{
			name:     "valid entry",
			entry:    IndexEntry{UrlString: "http://example.com", Count: 5},
			expected: `{"url_string":"http://example.com","count":5}`,
		},
		{
			name:     "empty entry",
			entry:    IndexEntry{},
			expected: `{"url_string":"","count":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Serialization
			data, err := json.Marshal(tt.entry)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}

			// Test Deserialization
			var unmarshaled IndexEntry
			err = json.Unmarshal([]byte(tt.expected), &unmarshaled)
			if err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if unmarshaled != tt.entry {
				t.Errorf("expected %+v, got %+v", tt.entry, unmarshaled)
			}
		})
	}
}
