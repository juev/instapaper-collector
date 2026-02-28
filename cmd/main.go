package main

import (
	"fmt"
	"os"
	"strconv"

	collector "github.com/juev/instapaper-collector"
	"github.com/juev/instapaper-collector/templates"
)

const storageFile = "data.json"

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

	userName := os.Getenv("USERNAME")
	if userName == "" {
		userName = "juev"
	}

	weekOffset := 47
	if v := os.Getenv("WEEK_OFFSET"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("WEEK_OFFSET must be a number: %w", err)
		}
		weekOffset = n
	}

	data := collector.New(storageFile)
	if err := data.Update(rssURL); err != nil {
		return err
	}

	return templates.TemplateFile(data, userName, weekOffset, ".")
}
