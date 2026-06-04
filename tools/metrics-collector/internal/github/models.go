package github

type Metrics struct {
	TotalRepos     int
	TotalStars     int
	TotalForks     int
	TotalWatchers  int
	TotalFollowers int
	TotalViews     int
	TopRepos       []TopRepo
}

type Repo struct {
	Name     string
	Stars    int
	Forks    int
	Watchers int
}

type TopRepo struct {
	Name  string
	Views int
}
