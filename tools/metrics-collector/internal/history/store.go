package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mhihasan/mhihasan/tools/metrics-collector/internal/github"
)

type Entry struct {
	Date         string `json:"date"`
	Repositories int    `json:"repositories"`
	Stars        int    `json:"stars"`
	Forks        int    `json:"forks"`
	Watchers     int    `json:"watchers"`
	Followers    int    `json:"followers"`
	Views14d     int    `json:"views_14d"`
}

type HistoryFile struct {
	Metrics []Entry `json:"metrics"`
}

// Update reads the history file at path, upserts today's entry, trims entries
// older than 365 days, and writes the result back. Returns error without
// modifying the file if the existing file contains corrupt JSON.
func Update(path string, date string, m github.Metrics) error {
	var h HistoryFile

	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read history file: %w", err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &h); err != nil {
			return fmt.Errorf("parse history file: %w", err)
		}
	}

	today := Entry{
		Date:         date,
		Repositories: m.TotalRepos,
		Stars:        m.TotalStars,
		Forks:        m.TotalForks,
		Watchers:     m.TotalWatchers,
		Followers:    m.TotalFollowers,
		Views14d:     m.TotalViews,
	}

	// Replace existing entry for the same date, or append
	found := false
	for i, e := range h.Metrics {
		if e.Date == date {
			h.Metrics[i] = today
			found = true
			break
		}
	}
	if !found {
		h.Metrics = append(h.Metrics, today)
	}

	// Trim entries older than 365 days relative to the date being written
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("parse date %q: %w", date, err)
	}
	cutoff := parsedDate.AddDate(0, 0, -365).Format("2006-01-02")
	kept := h.Metrics[:0]
	for _, e := range h.Metrics {
		if e.Date >= cutoff {
			kept = append(kept, e)
		}
	}
	h.Metrics = kept

	out, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	return os.WriteFile(path, out, 0644)
}
