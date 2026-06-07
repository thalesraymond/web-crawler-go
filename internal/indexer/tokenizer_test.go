package indexer

import (
	"reflect"
	"sort"
	"testing"
)

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
				{word: "hello", count: 1},
				{word: "world", count: 1},
				{word: "programming", count: 1},
			},
		},
		{
			name:     "stop words and short words",
			htmlBody: "the a os it hello a b c",
			want: []PageToken{
				{word: "hello", count: 1},
			},
		},
		{
			name:     "word counts",
			htmlBody: "test test hello test hello world",
			want: []PageToken{
				{word: "hello", count: 2},
				{word: "test", count: 3},
				{word: "world", count: 1},
			},
		},
		{
			name:     "html tags removed",
			htmlBody: "<html><body><h1>title</h1><p>paragraph text</p></body></html>",
			want: []PageToken{
				{word: "title", count: 1},
				{word: "paragraph", count: 1},
				{word: "text", count: 1},
			},
		},
		{
			name:     "ignored script and style",
			htmlBody: "<script>var x = 1;</script>hello<style>body { color: red; }</style>world",
			want: []PageToken{
				{word: "hello", count: 1},
				{word: "world", count: 1},
			},
		},
		{
			name:     "multiple ocurrences of a word",
			htmlBody: "<script>var x = 1;</script>hello<style>body { color: red; }</style><body><label>Here is the word</label><b>Again, the word!</b></body>",
			want: []PageToken{
				{word: "hello", count: 1},
				{word: "here", count: 1},
				{word: "word", count: 2},
				{word: "again", count: 1},
			},
		},
		{
			name:     "ignored nav, footer, header, aside",
			htmlBody: "<header>header_text</header><nav>nav_text</nav>main_content<footer>footer_text</footer><aside>aside_text</aside>",
			want: []PageToken{
				{word: "main_content", count: 1},
			},
		},
		{
			name:     "nested invalid tags",
			htmlBody: "<header><nav>SHOULDNOTBEHERE</nav>header_text</header><div><p>hello <span>world</span></p></div>",
			want: []PageToken{
				{word: "hello", count: 1},
				{word: "world", count: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPageTokens(tt.htmlBody)

			// Sort both slices to compare them, because map iteration order is random.
			sort.Slice(got, func(i, j int) bool {
				return got[i].word < got[j].word
			})

			wantCopy := make([]PageToken, len(tt.want))
			copy(wantCopy, tt.want)
			sort.Slice(wantCopy, func(i, j int) bool {
				return wantCopy[i].word < wantCopy[j].word
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
