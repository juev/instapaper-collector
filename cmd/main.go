package main

import (
	"fmt"
	"os"
	"strconv"

	collector "github.com/juev/instapaper-collector"
	"github.com/juev/instapaper-collector/templates"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	rssURL := os.Getenv("RSS_URL")
	if rssURL == "" {
		return fmt.Errorf("RSS_URL env variable is required")
	}

	userName := os.Getenv("GITHUB_USERNAME")
	if userName == "" {
		userName = "juev"
	}

	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "data.json"
	}

	weekOffset := 47
	if v := os.Getenv("WEEK_OFFSET"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("WEEK_OFFSET must be a number: %w", err)
		}
		weekOffset = n
	}

	data := collector.New(dataFile)
	if err := data.Update(rssURL); err != nil {
		return err
	}

	return templates.TemplateFile(data, userName, weekOffset, ".")
}
