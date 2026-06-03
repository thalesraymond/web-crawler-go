package main

import (
	"flag"
	"fmt"
	"os"
)

func runSearch(args []string) {
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)

	searchQuery := searchCmd.String("query", "", "Search query to search for")

	searchCmd.Parse(args)

	if *searchQuery == "" {
		fmt.Println("Error: Search query is required for search command")
		os.Exit(1)
	}

	fmt.Println("Searching for:", *searchQuery)
}
