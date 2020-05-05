package github

import (
	"github.com/m-zajac/goprojectdemo/internal/app"
)

type searchResponse struct {
	Items []searchResponseItem `json:"items"`
}

type searchResponseItem struct {
	ID    int                     `json:"id"`
	Name  string                  `json:"name"`
	Owner searchResponseItemOwner `json:"owner"`
}

type searchResponseItemOwner struct {
	Login string `json:"login"`
}

func (s searchResponse) ToProjects() []app.Project {
	ps := make([]app.Project, 0, len(s.Items))
	for _, i := range s.Items {
		ps = append(ps, app.Project{
			ID:         i.ID,
			Name:       i.Name,
			OwnerLogin: i.Owner.Login,
		})
	}

	return ps
}

type statsResponse []struct {
	Author statsResponseAuthor `json:"author"`
	Total  int                 `json:"total"`
}

type statsResponseAuthor struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

func (s statsResponse) ToStats() []app.ContributorStats {
	ss := make([]app.ContributorStats, 0, len(s))
	for _, el := range s {
		ss = append(ss, app.ContributorStats{
			Contributor: app.Contributor{
				ID:    el.Author.ID,
				Login: el.Author.Login,
			},
			Commits: el.Total,
		})
	}

	return ss
}
