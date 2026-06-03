package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected 'crawl' or 'search' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "crawl":
		runCrawl(os.Args[2:])
	case "search":
		runSearch(os.Args[2:])
	default:
		fmt.Println("Expected 'crawl' or 'search' subcommands")
		os.Exit(1)
	}

}
