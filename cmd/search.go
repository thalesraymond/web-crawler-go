package main

import (
	"flag"
	"fmt"
	"os"
)

func runSearch(args []string) {
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)

	searchQuery := searchCmd.String("query", "", "Search query to search for")

	_ = searchCmd.Parse(args) // Error handling is done by flag package, so we can ignore the error here

	if *searchQuery == "" {
		fmt.Println("Error: Search query is required for search command")
		os.Exit(1)
	}

	fmt.Println("Searching for:", *searchQuery)
}
