package internal

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/thalesraymond/web-crawler-go/internal/indexer"
	"github.com/thalesraymond/web-crawler-go/internal/network"
)

const MaxWorkers = 10

type Crawler struct {
	client        *network.CrawlerClient
	urlTracker    *network.URLTracker
	concurrency   int
	queue         chan string
	pageLimit     int
	currentResult chan *CrawlResult
	workerWg      sync.WaitGroup

	results []*CrawlResult
}

type CrawlResult struct {
	URL    string
	HTML   string
	Links  []string
	Tokens []indexer.PageToken
	Error  error
}

func (c *Crawler) GetResults() []*CrawlResult {
	return c.results
}

func NewCrawler(client *network.CrawlerClient, urlTracker *network.URLTracker, concurrency int, pageLimit int) *Crawler {
	return &Crawler{
		client:        client,
		urlTracker:    urlTracker,
		concurrency:   concurrency,
		queue:         make(chan string),
		pageLimit:     pageLimit,
		currentResult: make(chan *CrawlResult),
	}
}

func (c *Crawler) Start(seedUrl string) {
	log.Printf("Starting crawl at %s with %d workers", seedUrl, c.concurrency)

	c.currentResult = make(chan *CrawlResult, c.pageLimit)
	c.queue = make(chan string, c.pageLimit)

	c.workerWg.Add(c.concurrency)
	for i := 0; i < c.concurrency; i++ {
		go c.worker()
	}

	c.urlTracker.MarkVisited(seedUrl)
	c.queue <- seedUrl

	linksToProcess := 1 //Will prevent workers fron hanging if total links is less than page limit

	for linksToProcess > 0 && len(c.results) < c.pageLimit {
		linksToProcess--
		result := <-c.currentResult

		if result.Error != nil {
			continue
		}

		c.results = append(c.results, result)

		for _, link := range result.Links {
			if c.urlTracker.MarkVisited(link) {
				linksToProcess++

				// Do not block if queue is full
				go func(urlToSend string) {
					c.queue <- urlToSend
				}(link)
			}
		}
	}

	close(c.queue)
	close(c.currentResult)

}

func (c *Crawler) worker() {
	defer c.workerWg.Done()

	for url := range c.queue {
		log.Printf("Crawling URL: %s", url)
		c.crawlURL(url)
	}
}

func (c *Crawler) crawlURL(url string) {
	html, err := c.client.FetchHTML(context.Background(), url)

	time.Sleep(1000 * time.Millisecond)

	if err != nil {
		c.currentResult <- &CrawlResult{
			URL:   url,
			Error: err,
		}
		return
	}

	tokens := indexer.ExtractPageTokens(html)
	urls, err := network.ExtractLinks(url, html)

	if err != nil {
		c.currentResult <- &CrawlResult{
			URL:   url,
			Error: err,
		}

		return
	}

	c.currentResult <- &CrawlResult{
		URL:    url,
		HTML:   html,
		Links:  urls,
		Tokens: tokens,
		Error:  nil,
	}

}
