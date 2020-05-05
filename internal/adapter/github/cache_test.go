package github

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/m-zajac/goprojectdemo/internal/adapter/github/mock"
	"github.com/m-zajac/goprojectdemo/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCachedClientProjectsByLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		cacheSize      int
		callsWithCount []int
		callsInterval  time.Duration
		ttl            time.Duration
		wantErr        bool
		wantCalls      int
	}{
		{
			name:      "invalid cache size",
			cacheSize: 0,
			wantErr:   true,
		},
		{
			name:           "calls with same parameters",
			cacheSize:      1,
			callsWithCount: []int{2, 2, 2, 2},
			callsInterval:  time.Microsecond,
			ttl:            time.Minute,
			wantErr:        false,
			wantCalls:      1,
		},
		{
			name:           "some calls, then calls with smaller count param",
			cacheSize:      1,
			callsWithCount: []int{2, 2, 1, 1},
			callsInterval:  time.Microsecond,
			ttl:            time.Minute,
			wantErr:        false,
			wantCalls:      1,
		},
		{
			name:           "calls with various count params",
			cacheSize:      1,
			callsWithCount: []int{2, 2, 3, 3, 4, 5, 2, 2, 1},
			callsInterval:  time.Microsecond,
			ttl:            time.Minute,
			wantErr:        false,
			wantCalls:      4,
		},
		{
			name:           "calls with expiring ttl",
			cacheSize:      1,
			callsWithCount: []int{2, 2, 2, 2},
			callsInterval:  5 * time.Millisecond,
			ttl:            time.Millisecond,
			wantErr:        false,
			wantCalls:      4,
		},
	}

	projectsResponse := []app.Project{
		{
			ID:         1,
			Name:       "project1",
			OwnerLogin: "owner1",
		},
		{
			ID:         2,
			Name:       "project2",
			OwnerLogin: "owner2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var clientCalls int

			client := mock.NewMockGithubClient(ctrl)
			client.EXPECT().
				ProjectsByLanguage(gomock.Any(), "go", gomock.Any()).
				DoAndReturn(func(ctx context.Context, language string, count int) ([]app.Project, error) {
					clientCalls++
					return projectsResponse, nil
				}).
				AnyTimes()

			cachedClient, err := NewCachedClient(client, tt.cacheSize, tt.ttl)
			assert.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}

			for _, count := range tt.callsWithCount {
				projects, err := cachedClient.ProjectsByLanguage(context.Background(), "go", count)
				require.NoError(t, err)
				require.Equal(t, projectsResponse[0], projects[0])
				time.Sleep(tt.callsInterval)
			}

			assert.Equal(t, tt.wantCalls, clientCalls)
		})
	}
}

func TestCachedClientStatsByProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		cacheSize     int
		calls         int
		callsInterval time.Duration
		ttl           time.Duration
		wantErr       bool
		wantCalls     int
	}{
		{
			name:      "invalid cache size",
			cacheSize: 0,
			wantErr:   true,
		},
		{
			name:          "calls with same parameters",
			cacheSize:     1,
			calls:         4,
			callsInterval: time.Microsecond,
			ttl:           time.Minute,
			wantErr:       false,
			wantCalls:     1,
		},
		{
			name:          "calls with expiring ttl",
			cacheSize:     1,
			calls:         4,
			callsInterval: 5 * time.Millisecond,
			ttl:           time.Millisecond,
			wantErr:       false,
			wantCalls:     4,
		},
	}

	statsResponse := []app.ContributorStats{
		{
			Contributor: app.Contributor{
				ID:    1,
				Login: "person1",
			},
			Commits: 10,
		},
		{
			Contributor: app.Contributor{
				ID:    2,
				Login: "person2",
			},
			Commits: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var clientCalls int

			client := mock.NewMockGithubClient(ctrl)
			client.EXPECT().
				StatsByProject(gomock.Any(), "go", "golang").
				DoAndReturn(func(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
					clientCalls++
					return statsResponse, nil
				}).
				AnyTimes()

			cachedClient, err := NewCachedClient(client, tt.cacheSize, tt.ttl)
			assert.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}

			for i := 0; i < tt.calls; i++ {
				stats, err := cachedClient.StatsByProject(context.Background(), "go", "golang")
				require.NoError(t, err)
				require.Equal(t, statsResponse[0], stats[0])
				time.Sleep(tt.callsInterval)
			}

			assert.Equal(t, tt.wantCalls, clientCalls)
		})
	}
}
