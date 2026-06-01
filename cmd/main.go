package main

import (
	"fmt"
	"os"

	"github.com/thalesraymond/web-crawler-go/internal/indexer"
	"github.com/thalesraymond/web-crawler-go/internal/network"
)

func main() {
	fmt.Println("Hello World")

	net_error := network.Placeholder(os.Stdout)
	if net_error != nil {
		fmt.Println("Error: ", net_error)
		os.Exit(1)
	}

	indexer_error := indexer.Placeholder(os.Stdout)
	if indexer_error != nil {
		fmt.Println("Error: ", indexer_error)
		os.Exit(1)
	}
}
