package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// Client talks to the GitHub GraphQL and REST APIs.
type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

// NewClient creates a Client. baseURL defaults to https://api.github.com when empty.
func NewClient(token, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	return &Client{token: token, baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{}}
}

const graphQLQuery = `
query($username: String!, $cursor: String) {
  user(login: $username) {
    followers { totalCount }
    repositories(first: 100, after: $cursor, ownerAffiliations: OWNER, isFork: false, privacy: PUBLIC) {
      totalCount
      pageInfo { hasNextPage endCursor }
      nodes {
        name
        stargazerCount
        forkCount
        watchers { totalCount }
      }
    }
  }
}`

type graphQLResponse struct {
	Data struct {
		User struct {
			Followers struct {
				TotalCount int `json:"totalCount"`
			} `json:"followers"`
			Repositories struct {
				TotalCount int `json:"totalCount"`
				PageInfo   struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []struct {
					Name           string `json:"name"`
					StargazerCount int    `json:"stargazerCount"`
					ForkCount      int    `json:"forkCount"`
					Watchers       struct {
						TotalCount int `json:"totalCount"`
					} `json:"watchers"`
				} `json:"nodes"`
			} `json:"repositories"`
		} `json:"user"`
	} `json:"data"`
}

// CollectMetrics fetches all metrics for username and returns aggregated results.
func (c *Client) CollectMetrics(username string, topReposCount int) (Metrics, error) {
	var m Metrics

	// --- GraphQL: stars, forks, watchers, repos, followers ---
	firstPage := true
	cursor := ""
	for {
		vars := map[string]any{"username": username}
		if cursor != "" {
			vars["cursor"] = cursor
		}
		body, _ := json.Marshal(map[string]any{"query": graphQLQuery, "variables": vars})

		req, _ := http.NewRequest("POST", c.baseURL+"/graphql", bytes.NewReader(body))
		req.Header.Set("Authorization", "bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return m, fmt.Errorf("graphql request: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return m, fmt.Errorf("graphql returned HTTP %d", resp.StatusCode)
		}

		var gqlResp graphQLResponse
		if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
			resp.Body.Close()
			return m, fmt.Errorf("decode graphql response: %w", err)
		}
		resp.Body.Close()

		repos := gqlResp.Data.User.Repositories
		if firstPage {
			m.TotalFollowers = gqlResp.Data.User.Followers.TotalCount
			m.TotalRepos = repos.TotalCount
			firstPage = false
		}

		for _, n := range repos.Nodes {
			m.TotalStars += n.StargazerCount
			m.TotalForks += n.ForkCount
			m.TotalWatchers += n.Watchers.TotalCount
		}

		if !repos.PageInfo.HasNextPage || repos.PageInfo.EndCursor == "" {
			break
		}
		cursor = repos.PageInfo.EndCursor
	}

	// --- REST: traffic views ---
	type repoListItem struct {
		Name string `json:"name"`
	}
	type trafficResponse struct {
		Count int `json:"count"`
	}

	viewsByRepo := map[string]int{}
	page := 1
	for {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/users/%s/repos?per_page=100&page=%d&type=owner", c.baseURL, username, page), nil)
		req.Header.Set("Authorization", "token "+c.token)

		resp, err := c.http.Do(req)
		if err != nil {
			return m, fmt.Errorf("repos list request: %w", err)
		}
		var repoList []repoListItem
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err := json.Unmarshal(data, &repoList); err != nil || len(repoList) == 0 {
			break
		}

		for _, r := range repoList {
			vReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/repos/%s/%s/traffic/views", c.baseURL, username, r.Name), nil)
			vReq.Header.Set("Authorization", "token "+c.token)
			vResp, err := c.http.Do(vReq)
			if err != nil {
				continue
			}
			var tr trafficResponse
			json.NewDecoder(vResp.Body).Decode(&tr)
			vResp.Body.Close()
			if tr.Count > 0 {
				viewsByRepo[r.Name] = tr.Count
				m.TotalViews += tr.Count
			}
		}

		if len(repoList) < 100 {
			break
		}
		page++
	}

	// Build top repos list
	type kv struct {
		name  string
		views int
	}
	var sorted []kv
	for name, views := range viewsByRepo {
		sorted = append(sorted, kv{name, views})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].views > sorted[j].views })
	if len(sorted) > topReposCount {
		sorted = sorted[:topReposCount]
	}
	for _, item := range sorted {
		m.TopRepos = append(m.TopRepos, TopRepo{Name: item.name, Views: item.views})
	}

	return m, nil
}
