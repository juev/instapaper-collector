# instapaper-collector

[![build](https://github.com/juev/instapaper-collector/actions/workflows/build.yml/badge.svg)](https://github.com/juev/instapaper-collector/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/juev/instapaper-collector)](https://goreportcard.com/report/github.com/juev/instapaper-collector)

A CLI tool that collects links from an [Instapaper](https://www.instapaper.com/) RSS feed and publishes them as a weekly digest. The collected links are available at [juev/links](https://github.com/juev/links).

## Features

- Fetches and parses Instapaper RSS feeds
- Deduplicates links across runs
- Stores all collected items in a JSON file (`data.json`)
- Generates weekly Markdown digests grouped by ISO week
- Produces a `README.md` with the latest week's links

## Installation

### Binary

Download a pre-built binary from [Releases](https://github.com/juev/instapaper-collector/releases).

### Docker

```sh
docker pull ghcr.io/juev/instapaper-collector:latest
```

### From source

Requires Go 1.26+.

```sh
go install github.com/juev/instapaper-collector/cmd@latest
```

## Usage

```sh
export RSS_URL="https://www.instapaper.com/rss/..."
instapaper-collector
```

### Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `RSS_URL` | yes | — | Instapaper RSS feed URL |
| `DATA_FILE` | no | `data.json` | Path to the JSON data file |
| `GITHUB_USERNAME` | no | `juev` | Username for generated Markdown footer |
| `WEEK_OFFSET` | no | `47` | Hours to shift the ISO week boundary back from Monday 00:00 |

### Docker

```sh
docker run --rm \
  -e RSS_URL="https://www.instapaper.com/rss/..." \
  -v "$(pwd):/data" \
  -w /data \
  ghcr.io/juev/instapaper-collector:latest
```

## Development

```sh
# Run tests
go test -v -race ./...

# Lint
golangci-lint run ./...

# Build
go build -o instapaper-collector ./cmd/main.go
```

## License

MIT
