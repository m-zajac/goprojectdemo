package mock

import (
	"context"

	"github.com/m-zajac/goprojectdemo/app"
)

// GithubClient mocks app.GithubClient
type GithubClient struct {
	ProjectsByLanguageFunc func(ctx context.Context, language string, count int) ([]app.Project, error)
	StatsByProjectFunc     func(ctx context.Context, name string, owner string) ([]app.ContributorStats, error)
}

// ProjectsByLanguage returns projects by given programming language name
func (m *GithubClient) ProjectsByLanguage(ctx context.Context, language string, count int) ([]app.Project, error) {
	if m.ProjectsByLanguageFunc != nil {
		return m.ProjectsByLanguageFunc(ctx, language, count)
	}

	return []app.Project{}, nil
}

// StatsByProject returns stats by given github project params
func (m *GithubClient) StatsByProject(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
	if m.StatsByProjectFunc != nil {
		return m.StatsByProjectFunc(ctx, name, owner)
	}

	return []app.ContributorStats{}, nil
}
