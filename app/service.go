package app

import (
	"context"
	"sort"

	"github.com/pkg/errors"
)

// GithubClient returns details about gihub projects and stats
type GithubClient interface {
	ProjectsByLanguage(ctx context.Context, language string, count int) ([]Project, error)
	StatsByProject(ctx context.Context, name string, owner string) ([]ContributorStats, error)
}

// Service is main apps entry point. Provides all app functionality
type Service struct {
	githubClient GithubClient
}

// NewService creates new Service instance
func NewService(githubClient GithubClient) *Service {
	return &Service{
		githubClient: githubClient,
	}
}

// MostActiveContributors returns contributions with most commits.
// Contributions are taken from top `projectsCount` by the number of stars.
// Returns top `count` most active contributors by commit count.
func (s *Service) MostActiveContributors(
	ctx context.Context,
	language string,
	projectsCount int,
	count int,
) ([]ContributorStats, error) {
	if count <= 0 {
		return nil, errors.New("count must be greater than zero")
	}

	projects, err := s.githubClient.ProjectsByLanguage(ctx, language, projectsCount)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving projects")
	}

	type respWrapper struct {
		owner string
		name  string
		stats []ContributorStats
		err   error
	}
	responses := make(chan respWrapper, len(projects))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, p := range projects {
		p := p
		go func() {
			stats, err := s.githubClient.StatsByProject(ctx, p.Name, p.OwnerLogin)
			responses <- respWrapper{
				owner: p.OwnerLogin,
				name:  p.Name,
				stats: stats,
				err:   err,
			}
		}()
	}

	statsMap := make(map[int]ContributorStats)
	for i := 0; i < cap(responses); i++ {
		resp := <-responses
		if resp.err != nil {
			return nil, errors.Wrapf(resp.err, "retrievieng project %s/%s stats", resp.owner, resp.name)
		}

		for _, stat := range resp.stats {
			el, ok := statsMap[stat.Contributor.ID]
			if !ok {
				el = ContributorStats{
					Contributor: Contributor{
						ID:    stat.Contributor.ID,
						Login: stat.Contributor.Login,
					},
				}
			}
			el.Commits += stat.Commits
			statsMap[stat.Contributor.ID] = el
		}
	}

	result := make([]ContributorStats, 0, len(statsMap))
	for _, el := range statsMap {
		result = append(result, el)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Commits > result[j].Commits
	})

	if len(result) > count {
		result = result[:count]
	}

	return result, nil
}
