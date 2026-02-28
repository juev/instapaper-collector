package collector

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"
)

type Collector struct {
	Title    string `json:"title"`
	Updated  string `json:"updated"`
	Items    []Item `json:"items"`
	fileName string
	links    map[string]struct{}
}

type Item struct {
	Title       string `json:"title,omitempty"`
	Link        string `json:"link,omitempty"`
	Description string `json:"description,omitempty"`
	Published   string `json:"published,omitempty"`
}

func New(fileName string) *Collector {
	return &Collector{
		fileName: fileName,
		links:    make(map[string]struct{}),
	}
}

func (c *Collector) Read() error {
	if _, err := os.Stat(c.fileName); err != nil {
		return nil
	}

	data, err := os.ReadFile(c.fileName)
	if err != nil {
		return fmt.Errorf("cannot read file %q: %w", c.fileName, err)
	}

	if !json.Valid(data) {
		return nil
	}

	if err := json.Unmarshal(data, c); err != nil {
		return err
	}

	for _, item := range c.Items {
		c.links[item.Link] = struct{}{}
	}

	return nil
}

func (c *Collector) Write() error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "    ")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(c); err != nil {
		return err
	}

	if err := os.WriteFile(c.fileName, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("cannot create file %q: %w", c.fileName, err)
	}

	return nil
}

func (c *Collector) Update(rssURL string) error {
	if err := c.Read(); err != nil {
		return err
	}

	body, err := FetchRSS(rssURL)
	if err != nil {
		return err
	}

	items, err := ParseRSS(body)
	if err != nil {
		return err
	}

	for _, item := range items {
		if c.notContainsLink(item.Link) {
			c.Items = append(c.Items, item)
			c.links[item.Link] = struct{}{}
		}
	}

	slices.SortFunc(c.Items, func(a, b Item) int {
		return cmp.Compare(a.Published, b.Published)
	})

	c.Updated = time.Now().UTC().Format(time.RFC3339)

	return c.Write()
}

func FetchRSS(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS feed returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *Collector) notContainsLink(link string) bool {
	_, ok := c.links[link]
	return !ok
}
