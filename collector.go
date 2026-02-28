package collector

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"
)

const maxResponseSize = 10 << 20 // 10MB

var httpClient = &http.Client{Timeout: 30 * time.Second}

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
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("cannot stat file %q: %w", c.fileName, err)
	}

	data, err := os.ReadFile(c.fileName)
	if err != nil {
		return fmt.Errorf("cannot read file %q: %w", c.fileName, err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("invalid JSON in %q: %w", c.fileName, err)
	}

	filtered := c.Items[:0]
	for _, item := range c.Items {
		if item.Link == "" {
			continue
		}
		filtered = append(filtered, item)
		c.links[item.Link] = struct{}{}
	}
	c.Items = filtered

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

	tmp := c.fileName + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("cannot write temp file %q: %w", tmp, err)
	}

	if err := os.Rename(tmp, c.fileName); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("cannot rename %q to %q: %w", tmp, c.fileName, err)
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

	added := false
	for _, item := range items {
		if c.isNewLink(item.Link) {
			c.Items = append(c.Items, item)
			c.links[item.Link] = struct{}{}
			added = true
		}
	}

	if !added {
		return nil
	}

	slices.SortFunc(c.Items, func(a, b Item) int {
		return cmp.Compare(a.Published, b.Published)
	})

	c.Updated = time.Now().UTC().Format(time.RFC3339)

	return c.Write()
}

func FetchRSS(rawURL string) ([]byte, error) {
	resp, err := httpClient.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS feed returned status %d", resp.StatusCode)
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
}

func (c *Collector) isNewLink(link string) bool {
	_, ok := c.links[link]
	return !ok
}

// AbsFileName returns the absolute path of the data file (for testing).
func (c *Collector) AbsFileName() string {
	abs, err := filepath.Abs(c.fileName)
	if err != nil {
		return c.fileName
	}
	return abs
}
