package main

import (
	"flag"
	"fmt"
	"os"
)

func runCrawl(args []string) {
	crawlCmd := flag.NewFlagSet("crawl", flag.ExitOnError)
	seedUrl := crawlCmd.String("seed", "https://en.wikipedia.org/wiki/Main_Page", "Root URL to start crawling from")
	pageLimit := crawlCmd.Int("limit", 100, "Max number of pages to crawl")

	err := crawlCmd.Parse(args)

	if err != nil {
		fmt.Println("Error parsing crawl command arguments:", err)
		os.Exit(1)
	}

	if *seedUrl == "" {
		fmt.Println("Error: Seed URL is required for crawl command")
		os.Exit(1)
	}

	if *pageLimit <= 0 {
		fmt.Println("Error: Page limit must be greater than 0")
		os.Exit(1)
	}

	fmt.Println("Crawling website:", *seedUrl)
	fmt.Println("Max pages to crawl:", *pageLimit)
}
