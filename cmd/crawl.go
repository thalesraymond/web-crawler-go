package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thalesraymond/web-crawler-go/internal"
	"github.com/thalesraymond/web-crawler-go/internal/network"
)

func runCrawl(args []string) {
	crawlCmd := flag.NewFlagSet("crawl", flag.ExitOnError)
	seedUrl := crawlCmd.String("seed", "https://en.wikipedia.org/wiki/Main_Page", "Root URL to start crawling from")
	pageLimit := crawlCmd.Int("limit", 100, "Max number of pages to crawl")

	_ = crawlCmd.Parse(args) // Error handling is done by flag package, so we can ignore the error here

	if *seedUrl == "" {
		fmt.Println("Error: Seed URL is required for crawl command")
		os.Exit(1)
	}

	if *pageLimit <= 0 {
		fmt.Println("Error: Page limit must be greater than 0")
		os.Exit(1)
	}

	crawler := internal.NewCrawler(
		network.NewCrawlerClient(),
		network.NewURLTracker(),
		5,
		*pageLimit,
	)
	
	fmt.Println("Crawling website:", *seedUrl)
	fmt.Println("Max pages to crawl:", *pageLimit)

	crawler.Start(*seedUrl)

	results := crawler.GetResults()

	for i, result := range results {
		if result.Error != nil {
			fmt.Printf("%d. URL: %s\n", i+1, result.URL)
			fmt.Printf("   Error: %s\n", result.Error)
			continue
		}
		
		fmt.Printf("%d. URL: %s\n", i+1, result.URL)
		fmt.Printf("   Tokens: %d\n", len(result.Tokens))
		fmt.Printf("   Links: %d\n", len(result.Links))
		fmt.Println()
	}



}
