package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mhihasan/mhihasan/tools/metrics-collector/internal/github"
)

func TestCreatesNewFileWhenNoneExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics-history.json")

	err := Update(path, "2024-01-15", github.Metrics{TotalRepos: 3})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	var h HistoryFile
	if err := json.Unmarshal(data, &h); err != nil {
		t.Fatal(err)
	}
	if len(h.Metrics) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(h.Metrics))
	}
	if h.Metrics[0].Date != "2024-01-15" || h.Metrics[0].Repositories != 3 {
		t.Errorf("unexpected entry: %+v", h.Metrics[0])
	}
}

func TestTodaysEntryIsWrittenCorrectly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics-history.json")

	m := github.Metrics{
		TotalRepos:     10,
		TotalStars:     20,
		TotalForks:     5,
		TotalWatchers:  3,
		TotalFollowers: 50,
		TotalViews:     100,
	}
	if err := Update(path, "2024-06-01", m); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	var h HistoryFile
	json.Unmarshal(data, &h)

	e := h.Metrics[0]
	if e.Repositories != 10 || e.Stars != 20 || e.Forks != 5 ||
		e.Watchers != 3 || e.Followers != 50 || e.Views14d != 100 {
		t.Errorf("entry fields wrong: %+v", e)
	}
}

func TestDuplicateDateIsReplaced(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics-history.json")

	Update(path, "2024-06-01", github.Metrics{TotalStars: 5})
	Update(path, "2024-06-01", github.Metrics{TotalStars: 9})

	data, _ := os.ReadFile(path)
	var h HistoryFile
	json.Unmarshal(data, &h)

	if len(h.Metrics) != 1 {
		t.Fatalf("expected 1 entry after duplicate, got %d", len(h.Metrics))
	}
	if h.Metrics[0].Stars != 9 {
		t.Errorf("expected updated stars=9, got %d", h.Metrics[0].Stars)
	}
}

func TestEntriesOlderThan365DaysAreTrimmed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics-history.json")

	// Write an old entry directly
	old := HistoryFile{Metrics: []Entry{
		{Date: "2020-01-01", Repositories: 1},
		{Date: "2023-12-31", Repositories: 2},
	}}
	data, _ := json.Marshal(old)
	os.WriteFile(path, data, 0644)

	if err := Update(path, "2024-06-01", github.Metrics{TotalRepos: 5}); err != nil {
		t.Fatal(err)
	}

	data, _ = os.ReadFile(path)
	var h HistoryFile
	json.Unmarshal(data, &h)

	for _, e := range h.Metrics {
		if e.Date == "2020-01-01" {
			t.Error("entry from 2020 should have been trimmed")
		}
	}
}

func TestCorruptJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics-history.json")
	os.WriteFile(path, []byte("{corrupt json"), 0644)

	err := Update(path, "2024-06-01", github.Metrics{})
	if err == nil {
		t.Error("expected error for corrupt JSON")
	}

	// File must not be overwritten
	content, _ := os.ReadFile(path)
	if string(content) != "{corrupt json" {
		t.Error("corrupt file should not be overwritten")
	}
}
