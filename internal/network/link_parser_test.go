package network

import (
	"slices"
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
				if !slices.Contains(gotLinks, wantLink) {
					t.Errorf("expected link %q not found", wantLink)
				}
			}
		})
	}
}

func TestIsValidLink(t *testing.T) {
	tests := []struct {
		name string
		link string
		want bool
	}{
		{
			name: "valid http link",
			link: "http://example.com",
			want: true,
		},
		{
			name: "valid https link",
			link: "https://example.com",
			want: true,
		},
		{
			name: "valid relative link",
			link: "/path/to/page",
			want: true,
		},
		{
			name: "valid link without scheme",
			link: "example.com",
			want: true,
		},
		{
			name: "invalid javascript link",
			link: "javascript:void(0)",
			want: false,
		},
		{
			name: "invalid mailto link",
			link: "mailto:test@example.com",
			want: false,
		},
		{
			name: "invalid fragment link",
			link: "#section1",
			want: false,
		},
		{
			name: "invalid fragment only",
			link: "#",
			want: false,
		},
		{
			name: "invalid javascript mixed case",
			link: "JaVaScRiPt:void(0)",
			want: false,
		},
		{
			name: "invalid mailto mixed case",
			link: "MaIlTo:test@example.com",
			want: false,
		},
		{
			name: "empty link",
			link: "",
			want: true,
		},
		{
			name: "valid ftp link",
			link: "ftp://example.com/file",
			want: true,
		},
		{
			name: "valid ws link",
			link: "ws://example.com/socket",
			want: true,
		},
		{
			name: "invalid javascript uppercase",
			link: "JAVASCRIPT:alert(1)",
			want: false,
		},
		{
			name: "invalid mailto uppercase",
			link: "MAILTO:admin@localhost",
			want: false,
		},
		{
			name: "invalid fragment uppercase",
			link: "#SECTION",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidLink(tt.link); got != tt.want {
				t.Errorf("isValidLink() = %v, want %v", got, tt.want)
			}
		})
	}
}
