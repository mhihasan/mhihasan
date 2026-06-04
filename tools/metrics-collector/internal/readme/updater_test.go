package readme

import (
	"strings"
	"testing"

	"github.com/mhihasan/mhihasan/tools/metrics-collector/internal/github"
)

const sampleREADME = `# Profile

| Metric | Count |
|--------|-------|
| Repositories | <!--TOTAL_REPOS-->0<!--/TOTAL_REPOS--> |
| Stars | <!--TOTAL_STARS-->0<!--/TOTAL_STARS--> |
| Forks | <!--TOTAL_FORKS-->0<!--/TOTAL_FORKS--> |
| Watchers | <!--TOTAL_WATCHERS-->0<!--/TOTAL_WATCHERS--> |
| Followers | <!--TOTAL_FOLLOWERS-->0<!--/TOTAL_FOLLOWERS--> |
| Views (14 days) | <!--TOTAL_VIEWS-->0<!--/TOTAL_VIEWS--> |

<sub>Last updated: <!--LAST_UPDATED-->never<!--/LAST_UPDATED--></sub>

<!--TOP_REPOS_START-->
| Repository | Views |
|------------|-------|
<!--TOP_REPOS_END-->
`

func TestInjectsRepoCount(t *testing.T) {
	m := github.Metrics{TotalRepos: 12}
	out, err := Inject(sampleREADME, m, "2024-01-15")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<!--TOTAL_REPOS-->12<!--/TOTAL_REPOS-->") {
		t.Errorf("expected TOTAL_REPOS=12 in output, got:\n%s", out)
	}
}

func TestInjectsAllMetricFields(t *testing.T) {
	m := github.Metrics{
		TotalRepos:     5,
		TotalStars:     10,
		TotalForks:     3,
		TotalWatchers:  7,
		TotalFollowers: 99,
		TotalViews:     200,
	}
	out, err := Inject(sampleREADME, m, "2024-06-01")
	if err != nil {
		t.Fatal(err)
	}
	checks := []string{
		"<!--TOTAL_REPOS-->5<!--/TOTAL_REPOS-->",
		"<!--TOTAL_STARS-->10<!--/TOTAL_STARS-->",
		"<!--TOTAL_FORKS-->3<!--/TOTAL_FORKS-->",
		"<!--TOTAL_WATCHERS-->7<!--/TOTAL_WATCHERS-->",
		"<!--TOTAL_FOLLOWERS-->99<!--/TOTAL_FOLLOWERS-->",
		"<!--TOTAL_VIEWS-->200<!--/TOTAL_VIEWS-->",
		"<!--LAST_UPDATED-->2024-06-01<!--/LAST_UPDATED-->",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output", want)
		}
	}
}

func TestInjectsTopReposTable(t *testing.T) {
	m := github.Metrics{
		TopRepos: []github.TopRepo{
			{Name: "repo-a", Views: 50},
			{Name: "repo-b", Views: 30},
		},
	}
	out, err := Inject(sampleREADME, m, "2024-06-01")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "repo-a") || !strings.Contains(out, "repo-b") {
		t.Errorf("expected top repos table in output, got:\n%s", out)
	}
	if !strings.Contains(out, "| 50 |") || !strings.Contains(out, "| 30 |") {
		t.Errorf("expected view counts in top repos table")
	}
}

func TestOverwritesAlreadyInjectedValues(t *testing.T) {
	first, err := Inject(sampleREADME, github.Metrics{TotalStars: 5}, "2024-01-01")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Inject(first, github.Metrics{TotalStars: 8}, "2024-01-02")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(second, "<!--TOTAL_STARS-->5<!--/TOTAL_STARS-->") {
		t.Error("old value 5 should have been overwritten by 8")
	}
	if !strings.Contains(second, "<!--TOTAL_STARS-->8<!--/TOTAL_STARS-->") {
		t.Errorf("expected TOTAL_STARS=8, got:\n%s", second)
	}
}

func TestMissingPlaceholderReturnsError(t *testing.T) {
	_, err := Inject("# No placeholders here", github.Metrics{}, "2024-01-01")
	if err == nil {
		t.Error("expected error for README missing placeholders")
	}
}
