package collector

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRead_NonexistentFile(t *testing.T) {
	c := New(filepath.Join(t.TempDir(), "nonexistent.json"))

	if err := c.Read(); err != nil {
		t.Fatalf("Read() from nonexistent file should not error, got: %v", err)
	}

	if len(c.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(c.Items))
	}
}

func TestWriteRead_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	c := New(path)
	c.Title = "Test Title"
	c.Updated = "2025-02-28T10:00:00Z"
	c.Items = []Item{
		{
			Title:       "Article One",
			Link:        "https://example.com/one",
			Description: "First article description",
			Published:   "2025-02-28T09:00:00Z",
		},
		{
			Title:       "Article Two",
			Link:        "https://example.com/two",
			Description: "Second article description",
			Published:   "2025-02-28T10:00:00Z",
		},
	}

	if err := c.Write(); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	c2 := New(path)
	if err := c2.Read(); err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	if c2.Title != c.Title {
		t.Errorf("Title: got %q, want %q", c2.Title, c.Title)
	}
	if c2.Updated != c.Updated {
		t.Errorf("Updated: got %q, want %q", c2.Updated, c.Updated)
	}
	if len(c2.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(c2.Items))
	}
	if c2.Items[0].Title != "Article One" {
		t.Errorf("Items[0].Title: got %q, want %q", c2.Items[0].Title, "Article One")
	}
	if c2.Items[0].Description != "First article description" {
		t.Errorf("Items[0].Description: got %q, want %q", c2.Items[0].Description, "First article description")
	}
	if c2.Items[1].Link != "https://example.com/two" {
		t.Errorf("Items[1].Link: got %q, want %q", c2.Items[1].Link, "https://example.com/two")
	}
}

func TestRead_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	if err := os.WriteFile(path, []byte("{invalid json"), 0600); err != nil {
		t.Fatal(err)
	}

	c := New(path)
	if err := c.Read(); err == nil {
		t.Error("Read() should return error for invalid JSON")
	}
}

func TestDeduplication(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	c := New(path)
	c.Title = "Test"
	c.Items = []Item{
		{Title: "Existing", Link: "https://example.com/existing", Published: "2025-01-01T00:00:00Z"},
	}

	if err := c.Write(); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	c2 := New(path)
	if err := c2.Read(); err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	if !c2.isNewLink("https://example.com/new") {
		t.Error("isNewLink should return true for new link")
	}
	if c2.isNewLink("https://example.com/existing") {
		t.Error("isNewLink should return false for existing link")
	}
}

func TestWrite_JSONFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	c := New(path)
	c.Title = "Test"
	c.Items = []Item{
		{Title: "A & B", Link: "https://example.com/?a=1&b=2", Published: "2025-01-01T00:00:00Z"},
	}

	if err := c.Write(); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "\\u0026") {
		t.Error("JSON should not escape HTML entities (SetEscapeHTML(false))")
	}
}

func TestUpdate_WithHTTPServer(t *testing.T) {
	feedData, err := os.ReadFile("testdata/feed.xml")
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(feedData)
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "data.json")
	c := New(path)

	if err := c.Update(server.URL); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	if len(c.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(c.Items))
	}

	if c.Items[0].Published > c.Items[1].Published {
		t.Error("items should be sorted ascending by Published")
	}

	if c.Updated == "" {
		t.Error("Updated should be set after Update()")
	}

	if _, err := os.Stat(path); err != nil {
		t.Error("data.json should be written after Update()")
	}
}

func TestUpdate_Deduplication(t *testing.T) {
	feedData, err := os.ReadFile("testdata/feed.xml")
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(feedData)
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "data.json")
	c := New(path)

	if err := c.Update(server.URL); err != nil {
		t.Fatalf("first Update() error: %v", err)
	}

	firstCount := len(c.Items)

	c2 := New(path)
	if err := c2.Update(server.URL); err != nil {
		t.Fatalf("second Update() error: %v", err)
	}

	if len(c2.Items) != firstCount {
		t.Errorf("expected %d items after second Update (no duplicates), got %d", firstCount, len(c2.Items))
	}

	if c2.Updated != c.Updated {
		t.Error("Updated timestamp should not change when no new items added")
	}
}

