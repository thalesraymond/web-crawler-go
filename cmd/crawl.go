package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thalesraymond/web-crawler-go/internal"
	"github.com/thalesraymond/web-crawler-go/internal/network"
	"github.com/thalesraymond/web-crawler-go/internal/storage"
)

func runCrawl(args []string) {
	crawlCmd := flag.NewFlagSet("crawl", flag.ExitOnError)
	seedUrl := crawlCmd.String("seed", "https://en.wikipedia.org/wiki/Main_Page", "Root URL to start crawling from")
	pageLimit := crawlCmd.Int("limit", 5, "Max number of pages to crawl")

	_ = crawlCmd.Parse(args) // Error handling is done by flag package, so we can ignore the error here

	if *seedUrl == "" {
		fmt.Println("Error: Seed URL is required for crawl command")
		os.Exit(1)
	}

	if *pageLimit <= 0 {
		fmt.Println("Error: Page limit must be greater than 0")
		os.Exit(1)
	}

	// save to project directory data folder
	crawler := internal.NewCrawler(
		network.NewCrawlerClient(),
		network.NewURLTracker(),
		5,
		*pageLimit,
		storage.NewFileStorage("./data"),
	)

	fmt.Println("Crawling website:", *seedUrl)
	fmt.Println("Max pages to crawl:", *pageLimit)

	crawler.Start(*seedUrl)
}
