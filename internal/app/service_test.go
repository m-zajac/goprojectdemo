package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/m-zajac/goprojectdemo/internal/app"
	"github.com/m-zajac/goprojectdemo/internal/app/mock"
	"github.com/stretchr/testify/assert"
)

func TestServiceMostActiveContributors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMock     func(*mock.MockGithubClient)
		language      string
		projectsCount int
		count         int
		want          []app.ContributorStats
		wantErr       bool
	}{
		{
			name: "invalid count",
			setupMock: func(m *mock.MockGithubClient) {

			},
			language:      "go",
			projectsCount: 1,
			count:         0,
			want:          nil,
			wantErr:       true,
		},
		{
			name: "projects error from client",
			setupMock: func(m *mock.MockGithubClient) {
				m.EXPECT().
					ProjectsByLanguage(gomock.Any(), "go", 3).
					Return(nil, errors.New("error"))
			},
			language:      "go",
			projectsCount: 3,
			count:         1,
			want:          nil,
			wantErr:       true,
		},
		{
			name: "stats error from client",
			setupMock: func(m *mock.MockGithubClient) {
				m.EXPECT().
					ProjectsByLanguage(gomock.Any(), "go", 2).
					Return(
						[]app.Project{
							{
								ID:         1,
								Name:       "project",
								OwnerLogin: "owner",
							},
						},
						nil,
					)

				m.EXPECT().
					StatsByProject(gomock.Any(), "project", "owner").
					Return(nil, errors.New("error"))
			},
			language:      "go",
			projectsCount: 2,
			count:         1,
			want:          nil,
			wantErr:       true,
		},
		{
			name: "client ok, return valid, sorted response",
			setupMock: func(m *mock.MockGithubClient) {
				m.EXPECT().
					ProjectsByLanguage(gomock.Any(), "go", 1).
					Return(
						[]app.Project{
							{
								ID:         1,
								Name:       "project",
								OwnerLogin: "owner",
							},
						},
						nil,
					)

				m.EXPECT().
					StatsByProject(gomock.Any(), "project", "owner").
					Return(
						[]app.ContributorStats{
							{
								Commits: 3,
								Contributor: app.Contributor{
									ID:    1,
									Login: "cont1",
								},
							},
							{
								Commits: 5,
								Contributor: app.Contributor{
									ID:    2,
									Login: "cont2",
								},
							},
							{
								Commits: 4,
								Contributor: app.Contributor{
									ID:    3,
									Login: "cont3",
								},
							},
						},
						nil,
					)
			},
			language:      "go",
			projectsCount: 1,
			count:         2,
			want: []app.ContributorStats{
				{
					Commits: 5,
					Contributor: app.Contributor{
						ID:    2,
						Login: "cont2",
					},
				},
				{
					Commits: 4,
					Contributor: app.Contributor{
						ID:    3,
						Login: "cont3",
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

			githubCli := mock.NewMockGithubClient(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(githubCli)
			}

			s := app.NewService(githubCli, time.Minute)
			got, err := s.MostActiveContributors(
				context.Background(),
				tt.language,
				tt.projectsCount,
				tt.count,
			)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
