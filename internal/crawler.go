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

type ResultStorage interface {
	Save(result *CrawlResult) error
}

type Crawler struct {
	client        *network.CrawlerClient
	urlTracker    *network.URLTracker
	concurrency   int
	queue         chan string
	done          chan struct{}
	pageLimit     int
	crawlDelay    time.Duration
	storage       ResultStorage
	currentResult chan *CrawlResult
	workerWg      sync.WaitGroup
	sendWg        sync.WaitGroup
}

type CrawlResult struct {
	URL    string              `json:"url"`
	Links  []string            `json:"links"`
	Tokens []indexer.PageToken `json:"tokens"`
	Error  error               `json:"error"`
}

func NewCrawler(client *network.CrawlerClient, urlTracker *network.URLTracker, concurrency int, pageLimit int, storage ResultStorage) *Crawler {
	return &Crawler{
		client:      client,
		urlTracker:  urlTracker,
		concurrency: concurrency,
		pageLimit:   pageLimit,
		crawlDelay:  1000 * time.Millisecond,
		storage:     storage,
	}
}

func (c *Crawler) Start(seedUrl string) {
	log.Printf("Starting crawl at %s with %d workers", seedUrl, c.concurrency)

	c.currentResult = make(chan *CrawlResult, c.pageLimit)
	c.queue = make(chan string, c.pageLimit)
	c.done = make(chan struct{})

	totalCrawled := 0

	c.workerWg.Add(c.concurrency)
	for i := 0; i < c.concurrency; i++ {
		go c.worker()
	}

	c.urlTracker.MarkVisited(seedUrl)
	c.queue <- seedUrl

	linksToProcess := 1 // Will prevent workers from hanging if total links is less than page limit

	for linksToProcess > 0 && totalCrawled < c.pageLimit {
		linksToProcess--
		result := <-c.currentResult

		log.Printf("Result: %s", result.URL)

		if result.Error != nil {
			continue
		}

		totalCrawled++

		//save
		if err := c.storage.Save(result); err != nil {
			log.Printf("Error saving result: %v", err)
		}

		for _, link := range result.Links {
			if totalCrawled+linksToProcess >= c.pageLimit {
				break
			}

			if c.urlTracker.MarkVisited(link) {
				linksToProcess++

				// Send to queue without blocking; bail out if crawl is done.
				c.sendWg.Add(1)
				go func(urlToSend string) {
					defer c.sendWg.Done()
					select {
					case c.queue <- urlToSend:
					case <-c.done:
					}
				}(link)
			}
		}
	}

	// Signal all pending send goroutines to stop, then wait for them.
	close(c.done)
	c.sendWg.Wait()

	// No more sends; safe to close the queue so workers exit.
	close(c.queue)
	c.workerWg.Wait()
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

	time.Sleep(c.crawlDelay) // Pause to be gentle with the server and avoid rate limiting / ip ban

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
		Links:  urls,
		Tokens: tokens,
		Error:  nil,
	}

}
