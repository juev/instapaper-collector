package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	collector "github.com/juev/instapaper-collector"
)

func TestTemplateFile_BasicGeneration(t *testing.T) {
	dir := t.TempDir()

	c := &collector.Collector{
		Title:   "Instapaper: Unread",
		Updated: "2025-02-28T10:00:00Z",
		Items: []collector.Item{
			{Title: "Article One", Link: "https://example.com/one", Description: "First desc", Published: "2025-02-28T09:00:00Z"},
			{Title: "Article Two", Link: "https://example.com/two", Published: "2025-02-28T10:00:00Z"},
		},
	}

	if err := TemplateFile(c, "juev", 47, dir); err != nil {
		t.Fatalf("TemplateFile() error: %v", err)
	}

	readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatalf("README.md not created: %v", err)
	}

	content := string(readme)
	if !strings.Contains(content, "Article One") {
		t.Error("README.md should contain Article One")
	}
	if !strings.Contains(content, "https://example.com/one") {
		t.Error("README.md should contain link URL")
	}
	if !strings.Contains(content, "First desc") {
		t.Error("README.md should contain description")
	}
	if !strings.Contains(content, "/2 items)") {
		t.Error("README.md should show total count")
	}
}

func TestTemplateFile_MultipleWeeks(t *testing.T) {
	dir := t.TempDir()

	c := &collector.Collector{
		Title:   "Instapaper: Unread",
		Updated: "2025-03-10T10:00:00Z",
		Items: []collector.Item{
			{Title: "Week 9 Article", Link: "https://example.com/w9", Published: "2025-02-24T10:00:00Z"},
			{Title: "Week 10 Article", Link: "https://example.com/w10", Published: "2025-03-03T10:00:00Z"},
		},
	}

	if err := TemplateFile(c, "juev", 47, dir); err != nil {
		t.Fatalf("TemplateFile() error: %v", err)
	}

	dataDir := filepath.Join(dir, "data")
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatalf("data/ dir not created: %v", err)
	}

	if len(entries) < 2 {
		t.Errorf("expected at least 2 weekly files, got %d", len(entries))
	}

	readmeExists := false
	if _, err := os.Stat(filepath.Join(dir, "README.md")); err == nil {
		readmeExists = true
	}
	if !readmeExists {
		t.Error("README.md should exist")
	}
}

func TestTemplateFile_WeekOffsetZero(t *testing.T) {
	dir := t.TempDir()

	c := &collector.Collector{
		Title: "Test",
		Items: []collector.Item{
			{Title: "Monday Article", Link: "https://example.com/mon", Published: "2025-02-24T10:00:00Z"},
		},
	}

	if err := TemplateFile(c, "juev", 0, dir); err != nil {
		t.Fatalf("TemplateFile() error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatalf("data/ dir not created: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 weekly file, got %d", len(entries))
	}
}
