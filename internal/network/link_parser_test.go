package network

import (
	"testing"
)

func TestExtractLinks(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		htmlBody  string
		wantLinks []string
		wantErr   bool
	}{
		{
			name:    "simple page",
			baseURL: "https://example.com",
			htmlBody: `
			<html><body>
			<a href="/page1">Page 1</a>
			<a href="https://example.com/page2">Page 2</a>
			</body></html>
			`,
			wantLinks: []string{
				"https://example.com/page1",
				"https://example.com/page2",
			},
			wantErr: false,
		},
		{
			name:    "no links",
			baseURL: "https://example.com",
			htmlBody: `
			<html><body>
			</body></html>
			`,
			wantLinks: []string{},
			wantErr:   false,
		},
		{
			name:    "links with mixed case schemes",
			baseURL: "https://example.com",
			htmlBody: `
			<html><body>
			<a href="JavaScript:alert(1)">Mixed case Javascript link</a>
			<a href="MAILTO:test@example.com">Uppercase mailto link</a>
			<a href="/page1">Valid link</a>
			</body></html>
			`,
			wantLinks: []string{"https://example.com/page1"},
			wantErr:   false,
		},
		{
			name:    "invalid base url",
			baseURL: "",
			htmlBody: `
			<html><body>
			<a href="/page1">Page 1</a>
			</body></html>
			`,
			wantLinks: []string{},
			wantErr:   true,
		},
		{
			name:    "page with fragment",
			baseURL: "https://example.com",
			htmlBody: `
			<html><body>
			<a href="/page1#section1">Page 1</a>
			<a href="https://example.com/page2#section2">Page 2</a>
			<p> Hello! </p>
			<a href="https://example.com/page2#section2">Page 2</a>
			</body></html>
			`,
			wantLinks: []string{
				"https://example.com/page1",
				"https://example.com/page2",
				"https://example.com/page2",
			},
			wantErr: false,
		},
		{
			name:    "links with javascript or only frangment",
			baseURL: "https://example.com",
			htmlBody: `
			<html><body>
			<a href="javascript:void(0)">Javascript link</a>
			<a href="#section1">Fragment link</a>
			</body></html>
			`,
			wantLinks: []string{},
			wantErr:   false,
		},
		{
			name:    "invalid base url",
			baseURL: "",
			htmlBody: `
			<html><body>
			<a href="/page1">Page 1</a>
			</body></html>
			`,
			wantLinks: []string{},
			wantErr:   true,
		},
		{
			name:      "empty html body",
			baseURL:   "https://example.com",
			htmlBody:  "",
			wantLinks: []string{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks, err := ExtractLinks(tt.baseURL, tt.htmlBody)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
			if err != nil {
				return
			}
			if len(gotLinks) != len(tt.wantLinks) {
				t.Errorf("expected %d links, got %d", len(tt.wantLinks), len(gotLinks))
				return
			}

			for _, wantLink := range tt.wantLinks {
				if !contains(gotLinks, wantLink) {
					t.Errorf("expected link %q not found", wantLink)
				}
			}
		})
	}
}

func contains(links []string, link string) bool {
	for _, l := range links {
		if l == link {
			return true
		}
	}
	return false
}
