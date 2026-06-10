# 🕷️ web-crawler-go

A simple web crawler written in Go that simulates "Google-like" indexing. Built as the first personal project for the [Boot.dev](https://www.boot.dev/) course.

> **Status:** 🚧 Work in progress — the core crawling, indexing, and persistent storage are implemented. Search functionality is currently being built.

---

## Overview

`web-crawler-go` is a CLI tool with two subcommands:

| Command | Description |
|---------|-------------|
| `crawl` | Crawl a website starting from a seed URL, following links up to a configurable page limit. |
| `search` | Search the indexed pages for a given query. |

The goal is to build a fully functional crawler that fetches web pages, extracts links, and stores indexed content — then allows searching over it.

---

## Project Structure

```
web-crawler-go/
├── cmd/                        # Application entry point & CLI
│   ├── main.go                 # Main dispatcher (crawl / search subcommands)
│   ├── crawl.go                # crawl subcommand flags & logic
│   └── search.go               # search subcommand flags & logic
├── internal/
│   ├── crawler.go              # Core crawler logic (goroutines/channels)
│   ├── crawler_test.go
│   ├── network/                # HTTP fetching & link extraction
│   │   ├── http_client.go
│   │   ├── http_client_test.go
│   │   ├── link_parser.go
│   │   ├── link_parser_test.go
│   │   ├── url_tracker.go
│   │   └── url_tracker_test.go
│   ├── indexer/                # Page indexing & tokenization
│   │   ├── tokenizer.go
│   │   ├── tokenizer_test.go
│   │   ├── word_processor.go
│   │   └── word_processor_test.go
│   └── storage/                # Persistent storage
│       ├── file_storage.go
│       └── file_storage_test.go
├── bin/                        # Compiled binaries (git-ignored)
├── .github/workflows/ci.yml   # CI pipeline (build, test, lint)
├── go.mod
├── LICENSE                     # GPL-3.0
└── README.md
```

---

## Getting Started

### Prerequisites

- **Go 1.25.7+** — [install Go](https://go.dev/doc/install)

### Build

```bash
go build -o bin/crawler ./cmd/...
```

### Run

**Crawl a website:**

```bash
./bin/crawler crawl -seed "https://example.com" -limit 50
```

| Flag | Default | Description |
|------|---------|-------------|
| `-seed` | `https://en.wikipedia.org/wiki/Main_Page` | Root URL to start crawling from |
| `-limit` | `100` | Maximum number of pages to crawl |

**Search indexed pages:**

```bash
./bin/crawler search -query "golang concurrency"
```

| Flag | Default | Description |
|------|---------|-------------|
| `-query` | *(required)* | Search query string |

---

## Testing

Run all tests with the race detector enabled:

```bash
go test -v -race -cover ./...
```

---

## CI/CD

A GitHub Actions workflow runs on every push and pull request to `main`:

1. **Verify dependencies** — `go mod verify`
2. **Build** — compiles the binary to `bin/crawler`
3. **Test** — runs tests with race detection and coverage
4. **Lint** — runs `golangci-lint`
5. **Artifact upload** — stores the compiled binary

---

## Roadmap

- [x] Implement HTTP fetching in `internal/network`
- [x] Parse HTML and extract links
- [x] Implement BFS/DFS crawl strategy with depth & rate limiting
- [x] Build page indexer in `internal/indexer`
- [ ] Implement search over indexed content
- [x] Add persistent storage for crawl results
- [x] Add concurrency with goroutines & channels

---

## License

This project is licensed under the **GNU General Public License v3.0** — see the [LICENSE](LICENSE) file for details.
