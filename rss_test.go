package collector

import (
	"os"
	"testing"
)

func TestParseRSS_ValidFeed(t *testing.T) {
	data, err := os.ReadFile("testdata/feed.xml")
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}

	items, err := ParseRSS(data)
	if err != nil {
		t.Fatalf("ParseRSS() error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	if items[0].Title != "Article One" {
		t.Errorf("items[0].Title: got %q, want %q", items[0].Title, "Article One")
	}
	if items[0].Link != "https://example.com/article-one" {
		t.Errorf("items[0].Link: got %q, want %q", items[0].Link, "https://example.com/article-one")
	}
	if items[0].Description != "First article summary text." {
		t.Errorf("items[0].Description: got %q, want %q", items[0].Description, "First article summary text.")
	}
	if items[0].Published == "" {
		t.Error("items[0].Published should not be empty")
	}

	if items[1].Title != "Article Two" {
		t.Errorf("items[1].Title: got %q, want %q", items[1].Title, "Article Two")
	}
}

func TestParseRSS_EmptyFeed(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Instapaper: Unread</title>
<link>https://instapaper.com/u</link>
</channel>
</rss>`)

	items, err := ParseRSS(data)
	if err != nil {
		t.Fatalf("ParseRSS() error: %v", err)
	}

	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestParseRSS_NoTitle(t *testing.T) {
	data, err := os.ReadFile("testdata/feed.xml")
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}

	items, err := ParseRSS(data)
	if err != nil {
		t.Fatalf("ParseRSS() error: %v", err)
	}

	if items[2].Title != "Untitled" {
		t.Errorf("items[2].Title: got %q, want %q (empty title should become Untitled)", items[2].Title, "Untitled")
	}
}

func TestParseRSS_SkipsEmptyLink(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Test</title>
<item>
<title>Has Link</title>
<link>https://example.com/one</link>
<pubDate>Fri, 28 Feb 2025 10:00:00 GMT</pubDate>
</item>
<item>
<title>No Link</title>
<link></link>
<pubDate>Fri, 28 Feb 2025 09:00:00 GMT</pubDate>
</item>
</channel>
</rss>`)

	items, err := ParseRSS(data)
	if err != nil {
		t.Fatalf("ParseRSS() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item (empty link skipped), got %d", len(items))
	}

	if items[0].Title != "Has Link" {
		t.Errorf("items[0].Title: got %q, want %q", items[0].Title, "Has Link")
	}
}

func TestParseRSS_SkipsWhitespaceOnlyLink(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Test</title>
<item>
<title>Has Link</title>
<link>https://example.com/one</link>
<pubDate>Fri, 28 Feb 2025 10:00:00 GMT</pubDate>
</item>
<item>
<title>Whitespace Link</title>
<link>   </link>
<pubDate>Fri, 28 Feb 2025 09:00:00 GMT</pubDate>
</item>
</channel>
</rss>`)

	items, err := ParseRSS(data)
	if err != nil {
		t.Fatalf("ParseRSS() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item (whitespace link skipped), got %d", len(items))
	}

	if items[0].Title != "Has Link" {
		t.Errorf("items[0].Title: got %q, want %q", items[0].Title, "Has Link")
	}
}

func TestParseRSS_PubDateConversion(t *testing.T) {
	data, err := os.ReadFile("testdata/feed.xml")
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}

	items, err := ParseRSS(data)
	if err != nil {
		t.Fatalf("ParseRSS() error: %v", err)
	}

	want := "2025-02-28T10:00:00Z"
	if items[0].Published != want {
		t.Errorf("items[0].Published: got %q, want %q", items[0].Published, want)
	}
}
