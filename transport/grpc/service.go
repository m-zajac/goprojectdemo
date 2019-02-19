package grpc

import (
	"github.com/m-zajac/goprojectdemo/app"
	"github.com/pkg/errors"
	context "golang.org/x/net/context"
)

// AppService can return most active contributors.
type AppService interface {
	MostActiveContributors(
		ctx context.Context,
		language string,
		projectsCount int,
		count int,
	) ([]app.ContributorStats, error)
}

// Service implements ServiceServer definition, acting as a direct proxy to AppService.
type Service struct {
	appService AppService
}

// NewService returns new Service instance
func NewService(appService AppService) *Service {
	return &Service{
		appService: appService,
	}
}

// MostActiveContributors calls service and returns reply.
func (s *Service) MostActiveContributors(ctx context.Context, r *Request) (*Reply, error) {
	stats, err := s.appService.MostActiveContributors(
		ctx,
		r.Language,
		int(r.ProjectsCount),
		int(r.Count),
	)
	if err != nil {
		return nil, errors.Wrap(err, "service.MostActiveContributors")
	}

	replyStats := make([]*Stat, 0, len(stats))
	for _, st := range stats {
		replyStats = append(replyStats, &Stat{
			Contributor: &Contributor{
				Id:    int64(st.Contributor.ID),
				Login: st.Contributor.Login,
			},
			Commits: int32(st.Commits),
		})
	}
	return &Reply{
		Stat: replyStats,
	}, nil
}
