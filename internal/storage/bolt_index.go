package storage

import (
	"encoding/json"
	"errors"

	"github.com/thalesraymond/web-crawler-go/internal"
	"github.com/thalesraymond/web-crawler-go/internal/indexer"
	bolt "go.etcd.io/bbolt"
)

var bucketName = []byte("index")

// BoltIndex is a BoltDB-backed inverted index.
// It satisfies the IndexWriter and IndexReader interfaces
// defined at their respective consumer sites.
type BoltIndex struct {
	db       *bolt.DB
	filePath string
}

// LoadOrCreateBolt attempts to open an existing BoltDB index or creates a new one.
func LoadOrCreateBolt(filePath string) (*BoltIndex, error) {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &BoltIndex{db: db, filePath: filePath}, nil
}

// Add merges the tokens from a CrawlResult into the BoltDB index.
func (bi *BoltIndex) Add(result *internal.CrawlResult) error {
	if result == nil {
		return errors.New("crawl result cannot be nil")
	}
	if result.Error != nil {
		return result.Error
	}
	if len(result.Tokens) == 0 {
		return nil // nothing to index
	}

	return bi.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for _, token := range result.Tokens {
			var postings []internal.IndexEntry
			val := b.Get([]byte(token.Word))
			if val != nil {
				if err := json.Unmarshal(val, &postings); err != nil {
					return err
				}
			}

			updated := false
			for idx := range postings {
				if postings[idx].UrlString == result.URL {
					postings[idx].Count = token.Count // update in-place
					updated = true
					break
				}
			}

			if !updated {
				postings = append(postings, internal.IndexEntry{
					UrlString: result.URL,
					Count:     token.Count,
				})
			}

			newVal, err := json.Marshal(postings)
			if err != nil {
				return err
			}
			if err := b.Put([]byte(token.Word), newVal); err != nil {
				return err
			}
		}
		return nil
	})
}

// Save is a no-op for BoltIndex because bbolt persists to disk automatically on each transaction.
func (bi *BoltIndex) Save() error {
	return nil
}

// Close cleanly shuts down the BoltDB connection.
func (bi *BoltIndex) Close() error {
	if bi.db != nil {
		return bi.db.Close()
	}
	return nil
}

// Lookup returns the posting list for the given word.
func (bi *BoltIndex) Lookup(word string) ([]internal.IndexEntry, error) {
	normalized, err := indexer.ProcessWord(word)
	if err != nil {
		return nil, err
	}

	var result []internal.IndexEntry
	err = bi.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		val := b.Get([]byte(normalized))
		if val == nil {
			result = []internal.IndexEntry{}
			return nil
		}
		return json.Unmarshal(val, &result)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetRandomIndexedURL returns a randomly selected URL from the existing index.
func (bi *BoltIndex) GetRandomIndexedURL() (string, bool) {
	var randomURL string
	var found bool

	_ = bi.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		k, v := c.First()
		if k != nil && v != nil {
			var postings []internal.IndexEntry
			if err := json.Unmarshal(v, &postings); err == nil && len(postings) > 0 {
				randomURL = postings[0].UrlString
				found = true
			}
		}
		return nil
	})

	return randomURL, found
}
