package internal

// IndexEntry represents a single posting in the inverted index:
// the URL of a page and how many times a word appeared on it.
type IndexEntry struct {
	UrlString string `json:"url_string"`
	Count     int    `json:"count"`
}
