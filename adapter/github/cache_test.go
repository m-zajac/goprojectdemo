package github

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/m-zajac/goprojectdemo/app"
	"github.com/m-zajac/goprojectdemo/mock"
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
			var clientCalls int
			client := &mock.GithubClient{
				ProjectsByLanguageFunc: func(ctx context.Context, language string, count int) ([]app.Project, error) {
					clientCalls++
					return projectsResponse, nil
				},
			}
			cachedClient, err := NewCachedClient(client, tt.cacheSize, tt.ttl)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewCachedClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			for _, count := range tt.callsWithCount {
				projects, err := cachedClient.ProjectsByLanguage(context.Background(), "go", count)
				if err != nil {
					t.Fatalf("ProjectsByLanguage() subsequent call error = %v", err)
				}
				if !reflect.DeepEqual(projects[0], projectsResponse[0]) {
					t.Fatalf("ProjectsByLanguage() returned invalid first project: %+v, want %+v", projects[0], projectsResponse[0])
				}
				time.Sleep(tt.callsInterval)
			}

			if clientCalls != tt.wantCalls {
				t.Errorf("ProjectsByLanguage() called client %d times, want %d", clientCalls, tt.wantCalls)
			}
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
			var clientCalls int
			client := &mock.GithubClient{
				StatsByProjectFunc: func(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
					clientCalls++
					return statsResponse, nil
				},
			}
			cachedClient, err := NewCachedClient(client, tt.cacheSize, tt.ttl)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewCachedClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			for i := 0; i < tt.calls; i++ {
				stats, err := cachedClient.StatsByProject(context.Background(), "go", "golang")
				if err != nil {
					t.Fatalf("StatsByProject() subsequent call error = %v", err)
				}
				if !reflect.DeepEqual(stats[0], statsResponse[0]) {
					t.Fatalf("StatsByProject() returned invalid first project: %+v, want %+v", stats[0], statsResponse[0])
				}
				time.Sleep(tt.callsInterval)
			}

			if clientCalls != tt.wantCalls {
				t.Errorf("StatsByProject() called client %d times, want %d", clientCalls, tt.wantCalls)
			}
		})
	}
}
