package main

import (
	"fmt"
	"os"

	"github.com/thalesraymond/web-crawler-go/internal/indexer"
	"github.com/thalesraymond/web-crawler-go/internal/network"
)

func main() {
	fmt.Println("Hello World")
	network.Placeholder(os.Stdout)
	indexer.Placeholder(os.Stdout)
}
