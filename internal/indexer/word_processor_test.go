package indexer

import (
	"errors"
	"testing"
)

func TestProcessWord(t *testing.T) {
	tests := []struct {
		name        string
		rawWord     string
		wantedWord  string
		wantedError error
	}{
		{
			name:        "same input and output",
			rawWord:     "hello",
			wantedWord:  "hello",
			wantedError: nil,
		},
		{
			name:        "word at end of phrase",
			rawWord:     "hello.",
			wantedWord:  "hello",
			wantedError: nil,
		},
		{
			name:        "word with accents",
			rawWord:     "olá",
			wantedWord:  "ola",
			wantedError: nil,
		},
		{
			name:        "word with uppercase",
			rawWord:     "EITA",
			wantedWord:  "eita",
			wantedError: nil,
		},
		{
			name:        "multiple words (invalid, error)",
			rawWord:     "NOT A SINGLE WORD",
			wantedWord:  "",
			wantedError: errors.New("word contains more than one word"),
		},
		{
			name:        "word too short",
			rawWord:     "a",
			wantedWord:  "",
			wantedError: errors.New("word is too short"),
		},
		{
			name:        "stop word",
			rawWord:     "the",
			wantedWord:  "",
			wantedError: errors.New("word is a stop word"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessWord(tt.rawWord)

			if err != nil && err.Error() != tt.wantedError.Error() {
				t.Errorf("ProcessWord() error = %v, want %v", err, nil)
			}

			if got != tt.wantedWord {
				t.Errorf("ProcessWord() = %v, want %v", got, tt.wantedWord)
			}
		})
	}
}
