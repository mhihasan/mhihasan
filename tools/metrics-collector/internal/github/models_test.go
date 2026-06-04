package github

import "testing"

func TestMetricsZeroValue(t *testing.T) {
	var m Metrics
	if m.TotalRepos != 0 || m.TotalStars != 0 || m.TotalForks != 0 ||
		m.TotalWatchers != 0 || m.TotalFollowers != 0 || m.TotalViews != 0 {
		t.Error("zero Metrics should have all fields as 0")
	}
}

func TestRepoFields(t *testing.T) {
	r := Repo{Name: "my-repo", Stars: 5, Forks: 2, Watchers: 1}
	if r.Name != "my-repo" || r.Stars != 5 || r.Forks != 2 || r.Watchers != 1 {
		t.Error("Repo fields not set correctly")
	}
}

func TestTopRepoFields(t *testing.T) {
	tr := TopRepo{Name: "popular-repo", Views: 42}
	if tr.Name != "popular-repo" || tr.Views != 42 {
		t.Error("TopRepo fields not set correctly")
	}
}
