package grpc

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/m-zajac/goprojectdemo/app"

	"github.com/m-zajac/goprojectdemo/mock"
)

func TestServiceMostActiveContributors(t *testing.T) {
	tests := []struct {
		name           string
		req            Request
		appResultStats []app.ContributorStats
		appResultErr   error
		want           *Reply
		wantErr        bool
	}{
		{
			name: "app service error",
			req: Request{
				Language:      "x",
				ProjectsCount: 7,
				Count:         11,
			},
			appResultStats: nil,
			appResultErr:   errors.New("test error"),
			want:           nil,
			wantErr:        true,
		},
		{
			name: "app service ok, valid response",
			req: Request{
				Language:      "y",
				ProjectsCount: 13,
				Count:         2,
			},
			appResultStats: []app.ContributorStats{
				{
					Commits: 1,
					Contributor: app.Contributor{
						ID:    1,
						Login: "l1",
					},
				},
				{
					Commits: 2,
					Contributor: app.Contributor{
						ID:    5,
						Login: "l2",
					},
				},
			},
			appResultErr: nil,
			want: &Reply{
				Stat: []*Stat{
					{
						Commits: 1,
						Contributor: &Contributor{
							Id:    int64(1),
							Login: "l1",
						},
					},
					{
						Commits: 2,
						Contributor: &Contributor{
							Id:    int64(5),
							Login: "l2",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appService := &mock.Service{
				MostActiveContributorsFunc: func(
					ctx context.Context,
					language string,
					projectsCount int,
					count int,
				) ([]app.ContributorStats, error) {
					if language != tt.req.Language {
						t.Errorf("AppService.MostActiveContributors() got invalid language %s, want %s", language, tt.req.Language)
					}
					if projectsCount != int(tt.req.ProjectsCount) {
						t.Errorf("AppService.MostActiveContributors() got invalid projectsCount %d, want %d", projectsCount, tt.req.ProjectsCount)
					}
					if count != int(tt.req.Count) {
						t.Errorf("AppService.MostActiveContributors() got invalid count %d, want %d", count, tt.req.Count)
					}
					return tt.appResultStats, tt.appResultErr
				},
			}
			s := &Service{appService: appService}

			got, err := s.MostActiveContributors(context.Background(), &tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.MostActiveContributors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.MostActiveContributors() =\n\t%+v\nwant\n\t%+v", got, tt.want)
			}
		})
	}
}
