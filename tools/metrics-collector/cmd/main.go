package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mhihasan/mhihasan/tools/metrics-collector/internal/github"
	"github.com/mhihasan/mhihasan/tools/metrics-collector/internal/history"
	"github.com/mhihasan/mhihasan/tools/metrics-collector/internal/readme"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		return fmt.Errorf("GH_TOKEN environment variable is required")
	}
	username := os.Getenv("USERNAME")
	if username == "" {
		return fmt.Errorf("USERNAME environment variable is required")
	}

	readmePath := getenv("README_PATH", "README.md")
	historyPath := getenv("HISTORY_PATH", "docs/metrics-history.json")
	topReposCount, _ := strconv.Atoi(getenv("TOP_REPOS_COUNT", "10"))
	if topReposCount <= 0 {
		topReposCount = 10
	}

	client := github.NewClient(token, "")
	fmt.Printf("Collecting metrics for %s...\n", username)

	metrics, err := client.CollectMetrics(username, topReposCount)
	if err != nil {
		return fmt.Errorf("collect metrics: %w", err)
	}

	fmt.Printf("Repos: %d, Stars: %d, Forks: %d, Watchers: %d, Followers: %d, Views: %d\n",
		metrics.TotalRepos, metrics.TotalStars, metrics.TotalForks,
		metrics.TotalWatchers, metrics.TotalFollowers, metrics.TotalViews)

	today := time.Now().UTC().Format("2006-01-02")

	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("read README: %w", err)
	}
	updated, err := readme.Inject(string(readmeContent), metrics, today)
	if err != nil {
		return fmt.Errorf("inject README: %w", err)
	}
	if err := os.WriteFile(readmePath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("write README: %w", err)
	}
	fmt.Printf("README updated: %s\n", readmePath)

	if err := history.Update(historyPath, today, metrics); err != nil {
		return fmt.Errorf("update history: %w", err)
	}
	fmt.Printf("History updated: %s\n", historyPath)

	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
