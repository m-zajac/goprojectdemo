package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/m-zajac/goprojectdemo/internal/api/http/mock"
	"github.com/m-zajac/goprojectdemo/internal/app"
	"github.com/stretchr/testify/require"
)

func TestServiceMostActiveContributors(t *testing.T) {
	tests := []struct {
		name           string
		req            *Request
		appResultStats []app.ContributorStats
		appResultErr   error
		want           *Reply
		wantErr        bool
	}{
		{
			name: "app service error",
			req: &Request{
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
			req: &Request{
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			appService := mock.NewMockService(ctrl)
			appService.EXPECT().
				MostActiveContributors(gomock.Any(), tt.req.Language, int(tt.req.ProjectsCount), int(tt.req.Count)).
				Return(tt.appResultStats, tt.appResultErr)

			s := &Service{appService: appService}

			got, err := s.MostActiveContributors(context.Background(), tt.req)
			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.want, got)
		})
	}
}
