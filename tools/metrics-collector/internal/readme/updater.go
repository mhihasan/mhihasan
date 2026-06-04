package readme

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mhihasan/mhihasan/tools/metrics-collector/internal/github"
)

var requiredPlaceholders = []string{
	"TOTAL_REPOS", "TOTAL_STARS", "TOTAL_FORKS",
	"TOTAL_WATCHERS", "TOTAL_FOLLOWERS", "TOTAL_VIEWS",
	"LAST_UPDATED",
}

// Inject replaces all HTML-comment placeholders in content with values from m and date.
func Inject(content string, m github.Metrics, date string) (string, error) {
	for _, p := range requiredPlaceholders {
		if !strings.Contains(content, "<!--"+p+"-->") {
			return "", fmt.Errorf("README missing placeholder <!--%s-->", p)
		}
	}

	replacements := map[string]string{
		"TOTAL_REPOS":     fmt.Sprintf("%d", m.TotalRepos),
		"TOTAL_STARS":     fmt.Sprintf("%d", m.TotalStars),
		"TOTAL_FORKS":     fmt.Sprintf("%d", m.TotalForks),
		"TOTAL_WATCHERS":  fmt.Sprintf("%d", m.TotalWatchers),
		"TOTAL_FOLLOWERS": fmt.Sprintf("%d", m.TotalFollowers),
		"TOTAL_VIEWS":     fmt.Sprintf("%d", m.TotalViews),
		"LAST_UPDATED":    date,
	}

	for key, val := range replacements {
		re := regexp.MustCompile(`<!--` + key + `-->.*?<!--/` + key + `-->`)
		content = re.ReplaceAllString(content, "<!--"+key+"-->"+val+"<!--/"+key+"-->")
	}

	content = replaceTopRepos(content, m.TopRepos)
	return content, nil
}

func replaceTopRepos(content string, repos []github.TopRepo) string {
	var sb strings.Builder
	sb.WriteString("| Repository | Views |\n|------------|-------|\n")
	for _, r := range repos {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", r.Name, r.Views))
	}
	table := sb.String()

	re := regexp.MustCompile(`(?s)<!--TOP_REPOS_START-->.*?<!--TOP_REPOS_END-->`)
	return re.ReplaceAllString(content, "<!--TOP_REPOS_START-->\n"+table+"<!--TOP_REPOS_END-->")
}
