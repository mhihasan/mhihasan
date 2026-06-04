package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func singlePageServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphql" {
			resp := map[string]any{
				"data": map[string]any{
					"user": map[string]any{
						"followers": map[string]any{"totalCount": 42},
						"repositories": map[string]any{
							"totalCount": 2,
							"pageInfo":   map[string]any{"hasNextPage": false, "endCursor": ""},
							"nodes": []map[string]any{
								{"name": "repo-a", "stargazerCount": 10, "forkCount": 2, "watchers": map[string]any{"totalCount": 1}},
								{"name": "repo-b", "stargazerCount": 5, "forkCount": 1, "watchers": map[string]any{"totalCount": 3}},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		// Traffic views endpoint: return 0 repos so pagination stops
		if r.URL.Path == "/users/testuser/repos" {
			w.Write([]byte("[]"))
			return
		}
		http.NotFound(w, r)
	}))
}

func TestSinglePageReturnsCorrectTotals(t *testing.T) {
	srv := singlePageServer(t)
	defer srv.Close()

	c := NewClient("fake-token", srv.URL)
	m, err := c.CollectMetrics("testuser", 10)
	if err != nil {
		t.Fatal(err)
	}
	if m.TotalFollowers != 42 {
		t.Errorf("expected followers=42, got %d", m.TotalFollowers)
	}
	if m.TotalStars != 15 {
		t.Errorf("expected stars=15, got %d", m.TotalStars)
	}
	if m.TotalForks != 3 {
		t.Errorf("expected forks=3, got %d", m.TotalForks)
	}
	if m.TotalWatchers != 4 {
		t.Errorf("expected watchers=4, got %d", m.TotalWatchers)
	}
	if m.TotalRepos != 2 {
		t.Errorf("expected repos=2, got %d", m.TotalRepos)
	}
}

func TestMultiPageGraphQLPaginatesAndSums(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphql" {
			page++
			var resp map[string]any
			if page == 1 {
				resp = map[string]any{
					"data": map[string]any{
						"user": map[string]any{
							"followers": map[string]any{"totalCount": 10},
							"repositories": map[string]any{
								"totalCount": 3,
								"pageInfo":   map[string]any{"hasNextPage": true, "endCursor": "cursor1"},
								"nodes": []map[string]any{
									{"name": "r1", "stargazerCount": 3, "forkCount": 0, "watchers": map[string]any{"totalCount": 0}},
								},
							},
						},
					},
				}
			} else {
				resp = map[string]any{
					"data": map[string]any{
						"user": map[string]any{
							"followers": map[string]any{"totalCount": 10},
							"repositories": map[string]any{
								"totalCount": 3,
								"pageInfo":   map[string]any{"hasNextPage": false, "endCursor": ""},
								"nodes": []map[string]any{
									{"name": "r2", "stargazerCount": 7, "forkCount": 1, "watchers": map[string]any{"totalCount": 2}},
								},
							},
						},
					},
				}
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		if r.URL.Path == "/users/testuser/repos" {
			w.Write([]byte("[]"))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := NewClient("fake-token", srv.URL)
	m, err := c.CollectMetrics("testuser", 10)
	if err != nil {
		t.Fatal(err)
	}
	if m.TotalStars != 10 {
		t.Errorf("expected summed stars=10, got %d", m.TotalStars)
	}
	if m.TotalForks != 1 {
		t.Errorf("expected summed forks=1, got %d", m.TotalForks)
	}
}

func TestAPIErrorReturnsWrappedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewClient("bad-token", srv.URL)
	_, err := c.CollectMetrics("testuser", 10)
	if err == nil {
		t.Error("expected error for 401 response")
	}
}

func TestEmptyRepositoryListReturnsZeros(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphql" {
			resp := map[string]any{
				"data": map[string]any{
					"user": map[string]any{
						"followers": map[string]any{"totalCount": 0},
						"repositories": map[string]any{
							"totalCount": 0,
							"pageInfo":   map[string]any{"hasNextPage": false, "endCursor": ""},
							"nodes":      []map[string]any{},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		if r.URL.Path == "/users/testuser/repos" {
			w.Write([]byte("[]"))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := NewClient("fake-token", srv.URL)
	m, err := c.CollectMetrics("testuser", 10)
	if err != nil {
		t.Fatal(err)
	}
	if m.TotalStars != 0 || m.TotalForks != 0 || m.TotalWatchers != 0 || m.TotalFollowers != 0 {
		t.Errorf("expected all zeros for empty repo list, got %+v", m)
	}
}
