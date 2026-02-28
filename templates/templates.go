package templates

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	collector "github.com/juev/instapaper-collector"
)

//go:embed template.tmpl
var templateString string

type Data struct {
	Title    string
	UserName string
	Content  *collector.Collector
	Count    int
}

// TemplateFile generates weekly markdown files and README.md.
// weekOffset shifts the ISO week boundary BACK from Monday 00:00 by the given hours
// (e.g. 47 = Saturday 01:00, 0 = standard Monday).
func TemplateFile(s *collector.Collector, userName string, weekOffset int, baseDir string) error {
	tmpl, err := template.New("links").Parse(templateString)
	if err != nil {
		return err
	}

	var weekNumber, currentWeek string
	r := Data{UserName: userName}
	weekItems := &collector.Collector{Title: s.Title}
	offset := time.Duration(weekOffset) * time.Hour

	for _, item := range s.Items {
		t, err := time.Parse(time.RFC3339, item.Published)
		if err != nil {
			return err
		}

		year, week := t.Add(offset).ISOWeek()
		currentWeek = fmt.Sprintf("%d-%02d", year, week)

		if weekNumber != currentWeek {
			if weekNumber != "" {
				if err := writeTemplate(&r, weekNumber, weekItems, tmpl, baseDir); err != nil {
					return err
				}
			}
			weekNumber = currentWeek
			weekItems.Items = nil
		}
		weekItems.Items = append(weekItems.Items, item)
	}

	if weekNumber != "" {
		if err := writeTemplate(&r, weekNumber, weekItems, tmpl, baseDir); err != nil {
			return err
		}
	}

	r.Count = len(s.Items)
	lastWeek := &collector.Collector{Title: s.Title, Items: weekItems.Items}
	return writeTemplate(&r, "", lastWeek, tmpl, baseDir)
}

func writeTemplate(r *Data, weekNumber string, weekItems *collector.Collector, tmpl *template.Template, baseDir string) error {
	r.Title = weekNumber
	r.Content = weekItems
	fileName := filepath.Join(baseDir, "data", weekNumber+".md")

	if weekNumber == "" {
		r.Title = weekItems.Title
		fileName = filepath.Join(baseDir, "README.md")
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, r); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fileName), 0770); err != nil {
		return err
	}

	return os.WriteFile(fileName, []byte(buf.String()), 0644)
}
