package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/thalesraymond/web-crawler-go/internal"
	"github.com/thalesraymond/web-crawler-go/internal/network"
	"github.com/thalesraymond/web-crawler-go/internal/storage"
)

func runCrawl(args []string) {
	crawlCmd := flag.NewFlagSet("crawl", flag.ExitOnError)
	seedUrl := crawlCmd.String("seed", "", "Root URL to start crawling from. If omitted, uses random page from index or default Wikipedia page.")
	pageLimit := crawlCmd.Int("limit", 5, "Max number of pages to crawl")
	indexType := crawlCmd.String("index-type", "file", "Type of index to use: 'file' (in-memory JSON) or 'bolt' (BoltDB)")

	_ = crawlCmd.Parse(args) // Error handling is done by flag package, so we can ignore the error here

	if *pageLimit <= 0 {
		fmt.Println("Error: Page limit must be greater than 0")
		os.Exit(1)
	}

	// Define local interface to abstract the index implementation
	type AppIndex interface {
		internal.IndexWriter
		GetRandomIndexedURL() (string, bool)
	}

	var index AppIndex
	var err error

	if *indexType == "bolt" {
		index, err = storage.LoadOrCreateBolt("./data/index.db")
	} else {
		index, err = storage.LoadOrCreate("./data/index.json")
	}

	if err != nil {
		log.Fatalf("Error loading index: %v", err)
	}

	if *seedUrl == "" {
		url, ok := index.GetRandomIndexedURL()
		if ok {
			*seedUrl = url
			fmt.Println("No seed provided, using random page from index:", *seedUrl)
		} else {
			*seedUrl = "https://en.wikipedia.org/wiki/Main_Page"
			fmt.Println("No seed and no data provided, using default Wikipedia:", *seedUrl)
		}
	}

	// save to project directory data folder
	crawler := internal.NewCrawler(
		network.NewCrawlerClient(),
		network.NewURLTracker(),
		500,
		*pageLimit,
		storage.NewFileStorage("./data"),
		index,
	)

	fmt.Println("Crawling website:", *seedUrl)
	fmt.Println("Max pages to crawl:", *pageLimit)

	crawler.Start(*seedUrl)
}
