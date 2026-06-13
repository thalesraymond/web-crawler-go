package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/thalesraymond/web-crawler-go/internal"
	"github.com/thalesraymond/web-crawler-go/internal/storage"
)

// IndexReader is the read side of the inverted index.
// Defined here (at the consumer) so any backing implementation can be swapped.
type IndexReader interface {
	Lookup(word string) ([]internal.IndexEntry, error)
	Close() error
}

func runSearch(args []string) {
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchQuery := searchCmd.String("query", "", "Search query to search for")
	indexType := searchCmd.String("index-type", "file", "Type of index to use: 'file' (in-memory JSON) or 'bolt' (BoltDB)")
	_ = searchCmd.Parse(args)

	if *searchQuery == "" {
		fmt.Println("Error: Search query is required for search command")
		os.Exit(1)
	}

	var index IndexReader
	var err error

	if *indexType == "bolt" {
		index, err = storage.LoadOrCreateBolt("./data/index.db")
	} else {
		index, err = storage.LoadOrCreate("./data/index.json")
	}

	if err != nil {
		log.Fatalf("Error loading index: %v", err)
	}
	defer func() { _ = index.Close() }()

	results, err := index.Lookup(*searchQuery)
	if err != nil {
		fmt.Printf("No results: %v\n", err)
		os.Exit(0)
	}

	if len(results) == 0 {
		fmt.Printf("No pages found for %q\n", *searchQuery)
		os.Exit(0)
	}

	// Sort by count descending — highest term frequency first.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Count > results[j].Count
	})

	fmt.Printf("Results for %q (%d pages):\n\n", *searchQuery, len(results))
	for rank, entry := range results {
		fmt.Printf("  %d. %s (count: %d)\n", rank+1, entry.UrlString, entry.Count)
	}
}

