package mock

import (
	"context"

	"github.com/m-zajac/goprojectdemo/app"
)

// Service mocks app.Service
type Service struct {
	MostActiveContributorsFunc func(
		ctx context.Context,
		language string,
		projectsCount int,
		count int,
	) ([]app.ContributorStats, error)
}

// MostActiveContributors returns fake contributors data
func (s *Service) MostActiveContributors(
	ctx context.Context,
	language string,
	projectsCount int,
	count int,
) ([]app.ContributorStats, error) {
	if s.MostActiveContributorsFunc != nil {
		return s.MostActiveContributorsFunc(ctx, language, projectsCount, count)
	}

	return []app.ContributorStats{}, nil
}
