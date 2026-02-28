package collector

import (
	"encoding/xml"
	"fmt"
	"time"
)

type rss struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func ParseRSS(data []byte) ([]Item, error) {
	var feed rss
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse RSS: %w", err)
	}

	items := make([]Item, 0, len(feed.Channel.Items))
	for _, ri := range feed.Channel.Items {
		if ri.Link == "" {
			continue
		}

		title := ri.Title
		if title == "" {
			title = "Untitled"
		}

		published, err := parsePubDate(ri.PubDate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pubDate %q: %w", ri.PubDate, err)
		}

		items = append(items, Item{
			Title:       title,
			Link:        ri.Link,
			Description: ri.Description,
			Published:   published,
		})
	}

	return items, nil
}

func parsePubDate(s string) (string, error) {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t.UTC().Format(time.RFC3339), nil
		}
	}

	return "", fmt.Errorf("unsupported date format: %s", s)
}
