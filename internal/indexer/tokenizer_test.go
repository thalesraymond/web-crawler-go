package indexer

import (
	"reflect"
	"sort"
	"testing"
)

func TestIsInvalidTag(t *testing.T) {
	tests := []struct {
		name    string
		tagName string
		want    bool
	}{
		{"script tag", "script", true},
		{"style tag", "style", true},
		{"nav tag", "nav", true},
		{"footer tag", "footer", true},
		{"header tag", "header", true},
		{"aside tag", "aside", true},
		{"div tag", "div", false},
		{"p tag", "p", false},
		{"span tag", "span", false},
		{"empty string", "", false},
		{"unrecognized tag", "customtag", false},
		{"body tag", "body", false},
		{"head tag", "head", false},
		{"b tag", "b", false},
		{"strong tag", "strong", false},
		{"i tag", "i", false},
		{"em tag", "em", false},
		{"uppercase script tag", "SCRIPT", true},
		{"uppercase style tag", "STYLE", true},
		{"uppercase nav tag", "NAV", true},
		{"uppercase footer tag", "FOOTER", true},
		{"uppercase header tag", "HEADER", true},
		{"uppercase aside tag", "ASIDE", true},
		{"mixedcase script tag", "sCrIpT", true},
		{"mixedcase style tag", "StYlE", true},
		{"mixedcase nav tag", "NaV", true},
		{"mixedcase footer tag", "fOoTeR", true},
		{"mixedcase header tag", "HeAdEr", true},
		{"mixedcase aside tag", "aSiDe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isInvalidTag(tt.tagName); got != tt.want {
				t.Errorf("isInvalidTag(%q) = %v, want %v", tt.tagName, got, tt.want)
			}
		})
	}
}

func FuzzIsInvalidTag(f *testing.F) {
	// Add seed corpus for fuzzer
	seedCorpus := []string{
		"script", "STYLE", "nav", "fOoTeR", "header", "ASIDE",
		"div", "p", "span", "", "customtag", "SCRIPT",
	}
	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, tagName string) {
		// The goal of the fuzz test is to ensure isInvalidTag doesn't panic
		// and consistently returns a bool for any random string.
		_ = isInvalidTag(tagName)
	})
}

func TestExtractPageTokens(t *testing.T) {
	tests := []struct {
		name     string
		htmlBody string
		want     []PageToken
	}{
		{
			name:     "empty html",
			htmlBody: "",
			want:     []PageToken{},
		},
		{
			name:     "plain text without stop words",
			htmlBody: "hello world programming",
			want: []PageToken{
				{Word: "hello", Count: 1},
				{Word: "world", Count: 1},
				{Word: "programming", Count: 1},
			},
		},
		{
			name:     "stop words and short words",
			htmlBody: "the a os it hello a b c",
			want: []PageToken{
				{Word: "hello", Count: 1},
			},
		},
		{
			name:     "word counts",
			htmlBody: "test test hello test hello world",
			want: []PageToken{
				{Word: "hello", Count: 2},
				{Word: "test", Count: 3},
				{Word: "world", Count: 1},
			},
		},
		{
			name:     "html tags removed",
			htmlBody: "<html><body><h1>title</h1><p>paragraph text</p></body></html>",
			want: []PageToken{
				{Word: "title", Count: 1},
				{Word: "paragraph", Count: 1},
				{Word: "text", Count: 1},
			},
		},
		{
			name:     "ignored script and style",
			htmlBody: "<script>var x = 1;</script>hello<style>body { color: red; }</style>world",
			want: []PageToken{
				{Word: "hello", Count: 1},
				{Word: "world", Count: 1},
			},
		},
		{
			name:     "multiple ocurrences of a word",
			htmlBody: "<script>var x = 1;</script>hello<style>body { color: red; }</style><body><label>Here is the word</label><b>Again, the word!</b></body>",
			want: []PageToken{
				{Word: "hello", Count: 1},
				{Word: "here", Count: 1},
				{Word: "word", Count: 2},
				{Word: "again", Count: 1},
			},
		},
		{
			name:     "case sensitive invalid tags",
			htmlBody: "<SCRIPT>var x = 1;</SCRIPT>hello<STYLE>body { color: red; }</STYLE>world<NaV>nav_text</nAv>!",
			want: []PageToken{
				{Word: "hello", Count: 1},
				{Word: "world", Count: 1},
			},
		},
		{
			name:     "ignored nav, footer, header, aside",
			htmlBody: "<header>header_text</header><nav>nav_text</nav>main_content<footer>footer_text</footer><aside>aside_text</aside>",
			want: []PageToken{
				{Word: "main_content", Count: 1},
			},
		},
		{
			name:     "nested invalid tags",
			htmlBody: "<header><nav>SHOULDNOTBEHERE</nav>header_text</header><div><p>hello <span>world</span></p></div>",
			want: []PageToken{
				{Word: "hello", Count: 1},
				{Word: "world", Count: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPageTokens(tt.htmlBody)

			// Sort both slices to compare them, because map iteration order is random.
			sort.Slice(got, func(i, j int) bool {
				return got[i].Word < got[j].Word
			})

			wantCopy := make([]PageToken, len(tt.want))
			copy(wantCopy, tt.want)
			sort.Slice(wantCopy, func(i, j int) bool {
				return wantCopy[i].Word < wantCopy[j].Word
			})

			if len(got) == 0 && len(wantCopy) == 0 {
				return
			}

			if !reflect.DeepEqual(got, wantCopy) {
				t.Errorf("ExtractPageTokens() = %v, want %v", got, wantCopy)
			}
		})
	}
}
