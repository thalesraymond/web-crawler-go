package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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

	httpClient := network.NewCrawlerClient()

	// testing the HTTP client by fetching the HTML content of the seed URL
	// TODO: This is just to test a real world download case, will be removed after the
	// real crawling logic is implemented
	ctx := context.Background()
	
	html, err := httpClient.FetchHTML(ctx, *seedUrl)
	if err != nil {
		fmt.Println("Error fetching HTML:", err)
		os.Exit(1)
	}
	
	fmt.Println("Fetched HTML content of length:", len(html))
	fmt.Println("Crawling website:", *seedUrl)
	fmt.Println("Max pages to crawl:", *pageLimit)
}
