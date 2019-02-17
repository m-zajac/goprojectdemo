package app_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/m-zajac/goprojectdemo/app"
	"github.com/m-zajac/goprojectdemo/mock"
	"github.com/pkg/errors"
)

func TestServiceMostActiveContributors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		newGithubClient func(*testing.T) *mock.GithubClient
		language        string
		projectsCount   int
		count           int
		want            []app.ContributorStats
		wantErr         bool
	}{
		{
			name: "invalid count",
			newGithubClient: func(*testing.T) *mock.GithubClient {
				return &mock.GithubClient{
					ProjectsByLanguageFunc: func(ctx context.Context, language string, count int) ([]app.Project, error) {
						t.Fatal("unwanted call for ProjectsByLanguage")
						return nil, nil
					},
					StatsByProjectFunc: func(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
						t.Fatal("unwanted call for StatsByProject")
						return nil, nil
					},
				}
			},
			language:      "go",
			projectsCount: 1,
			count:         0,
			want:          nil,
			wantErr:       true,
		},
		{
			name: "projects error from client",
			newGithubClient: func(*testing.T) *mock.GithubClient {
				return &mock.GithubClient{
					ProjectsByLanguageFunc: func(ctx context.Context, language string, count int) ([]app.Project, error) {
						if count != 3 {
							t.Errorf("invalid count arg, want 3, got %d", count)
						}
						return nil, errors.New("error")
					},
				}
			},
			language:      "go",
			projectsCount: 3,
			count:         1,
			want:          nil,
			wantErr:       true,
		},
		{
			name: "stats error from client",
			newGithubClient: func(*testing.T) *mock.GithubClient {
				return &mock.GithubClient{
					ProjectsByLanguageFunc: func(ctx context.Context, language string, count int) ([]app.Project, error) {
						if count != 2 {
							t.Errorf("invalid count arg, want 2, got %d", count)
						}
						return []app.Project{
							{
								ID:         1,
								Name:       "project",
								OwnerLogin: "owner",
							},
						}, nil
					},
					StatsByProjectFunc: func(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
						return nil, errors.New("error")
					},
				}
			},
			language:      "go",
			projectsCount: 2,
			count:         1,
			want:          nil,
			wantErr:       true,
		},
		{
			name: "client ok, return valid, sorted response",
			newGithubClient: func(*testing.T) *mock.GithubClient {
				return &mock.GithubClient{
					ProjectsByLanguageFunc: func(ctx context.Context, language string, count int) ([]app.Project, error) {
						if count != 1 {
							t.Errorf("invalid count arg, want 1, got %d", count)
						}
						return []app.Project{
							{
								ID:         1,
								Name:       "project",
								OwnerLogin: "owner",
							},
						}, nil
					},
					StatsByProjectFunc: func(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
						if name != "project" {
							t.Errorf("invalid name arg, want 'project', got %s", name)
						}
						if owner != "owner" {
							t.Errorf("invalid owner arg, want 'owner', got %s", name)
						}
						return []app.ContributorStats{
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
						}, nil
					},
				}
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
			s := app.NewService(tt.newGithubClient(t))
			got, err := s.MostActiveContributors(
				context.Background(),
				tt.language,
				tt.projectsCount,
				tt.count,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.MostActiveContributors() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.MostActiveContributors() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
