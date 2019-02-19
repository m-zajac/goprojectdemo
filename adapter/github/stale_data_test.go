package github

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/m-zajac/goprojectdemo/app"
	"github.com/m-zajac/goprojectdemo/mock"
	"github.com/sirupsen/logrus"
)

// TestClientWithStaleDataScheduler test scheduler.
// It's a form of white box test - every scheduler step is checked one by one.
// This code is a little dirty. Testing concurent code is hard.
func TestClientWithStaleDataScheduler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		newStaleDataClientCall func(*ClientWithStaleData) func() error
	}{
		{
			name: "ProjectsByLanguage",
			newStaleDataClientCall: func(c *ClientWithStaleData) func() error {
				return func() error {
					_, err := c.ProjectsByLanguage(context.Background(), "go", 1)
					return err
				}
			},
		},
		{
			name: "StatsByProject",
			newStaleDataClientCall: func(c *ClientWithStaleData) func() error {
				return func() error {
					_, err := c.StatsByProject(context.Background(), "golang", "go")
					return err
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var _clientCalls int64
			getClientCalls := func() int {
				v := atomic.LoadInt64(&_clientCalls)
				return int(v)
			}
			clientTokens := make(chan struct{}, 1)
			client := &mock.GithubClient{
				ProjectsByLanguageFunc: func(ctx context.Context, language string, count int) ([]app.Project, error) {
					select {
					case <-clientTokens:
					case <-time.After(time.Second):
						t.Fatal("client locked")
					}

					atomic.AddInt64(&_clientCalls, int64(1))

					return nil, nil
				},
				StatsByProjectFunc: func(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
					select {
					case <-clientTokens:
					case <-time.After(time.Second):
						t.Fatal("client locked")
					}

					atomic.AddInt64(&_clientCalls, int64(1))

					return nil, nil
				},
			}
			storeTokens := make(chan struct{}, 10)
			store := mock.NewKVStore(nil, storeTokens)
			l := logrus.New()

			staleDataClient, err := NewClientWithStaleData(client, store, time.Minute, l)
			if err != nil {
				t.Fatalf("NewClientWithStaleData() error = %v", err)
			}

			// Set special chan for blocking scheduler
			staleDataClient.schedulerPendingOps = make(chan int, 1)

			staleDataClient.RunScheduler()

			staleDataClientCall := tt.newStaleDataClientCall(staleDataClient)

			pendingUpdates := 0
			expectedClientCalls := 0
			expectedStoreReads := 0
			expectedStoreUpdates := 0
			expectedPendingUpdates := 0
			checkNextState := func(step string) {
				select {
				case pendingUpdates = <-staleDataClient.schedulerPendingOps:
				case <-time.After(time.Second):
					t.Fatalf("%s: scheduler locked", step)
				}

				time.Sleep(10 * time.Millisecond)

				if pendingUpdates != expectedPendingUpdates {
					t.Errorf("%s: pending scheduler updates count invalid: %d, want %d", step, pendingUpdates, expectedPendingUpdates)
				}
				if v := getClientCalls(); v != expectedClientCalls {
					t.Errorf("%s: client call count invalid: %d, want %d", step, v, expectedClientCalls)
				}
				if v := store.Updates(); v != expectedStoreUpdates {
					t.Errorf("%s: store update count invalid: %d, want %d", step, v, expectedStoreUpdates)
				}
				if v := store.Reads(); v != expectedStoreReads {
					t.Errorf("%s: store read count invalid: %d, want %d", step, v, expectedStoreReads)
				}
			}

			checkNextState("init scheduler")

			// PHASE1: Read with empty db
			t.Log("PHASE1: First call - should read from db, schedule update then call client")
			expectedClientCalls++
			clientTokens <- struct{}{} // allow client call
			if err = staleDataClientCall(); err != nil {
				t.Errorf("phase1: ClientWithStaleData call error = %v", err)
			}
			expectedStoreReads++
			expectedPendingUpdates++
			checkNextState("phase1: after ClientWithStaleData call")

			t.Log("PHASE1: Next scheduler state - should see empty pending queue and store update")
			expectedPendingUpdates--
			expectedStoreUpdates++
			storeTokens <- struct{}{} // allow store write
			checkNextState("phase1: after scheduler finishes updates")

			// PHASE2: Read with data already in db
			t.Log("PHASE2: Second call - should read from db but NOT call client, should schedule update")
			if err = staleDataClientCall(); err != nil {
				t.Errorf("phase2: ClientWithStaleData call error = %v", err)
			}
			expectedStoreReads++
			expectedPendingUpdates++
			checkNextState("phase2: after ClientWithStaleData call")

			t.Log("PHASE2: Next scheduler state - should see client called, empty pending queue and new store update")
			expectedPendingUpdates--
			expectedStoreUpdates++
			storeTokens <- struct{}{} // allow store write
			expectedClientCalls++
			clientTokens <- struct{}{} // allow client call
			checkNextState("phase2: after scheduler finishes updates")

			// PHASE3: Read with data in db, but ttl exceeded
			t.Log("PHASE3: Third call - should read from db, schedule update then call client")
			expectedClientCalls++
			clientTokens <- struct{}{} // allow client call
			if err = staleDataClientCall(); err != nil {
				t.Errorf("phase3: ClientWithStaleData call error = %v", err)
			}
			expectedStoreReads++
			expectedPendingUpdates++
			checkNextState("phase3: after ClientWithStaleData call")

			t.Log("PHASE3: Next scheduler state - should see empty pending queue and store update")
			expectedPendingUpdates--
			expectedStoreUpdates++
			storeTokens <- struct{}{} // allow store write
			checkNextState("phase3: after scheduler finishes updates")
		})
	}
}

func TestClientWithStaleDataProjectsByLanguage(t *testing.T) {
	t.Parallel()

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

	var clientCalls int
	client := &mock.GithubClient{
		ProjectsByLanguageFunc: func(ctx context.Context, language string, count int) ([]app.Project, error) {
			clientCalls++
			return projectsResponse, nil
		},
	}
	store := mock.NewKVStore(nil, nil)
	l := logrus.New()

	staleDataClient, err := NewClientWithStaleData(client, store, time.Minute, l)
	if err != nil {
		t.Fatalf("NewClientWithStaleData() error = %v", err)
	}

	projects, err := staleDataClient.ProjectsByLanguage(context.Background(), "go", 2)
	if err != nil {
		t.Fatalf("ProjectsByLanguage() subsequent call error = %v", err)
	}
	if !reflect.DeepEqual(projects, projectsResponse) {
		t.Fatalf("ProjectsByLanguage() returned invalid first project: %+v, want %+v", projects, projectsResponse)
	}
	if clientCalls != 1 {
		t.Errorf("ProjectsByLanguage() called client %d times, want %d", clientCalls, 1)
	}
}

func TestClientWithStaleDataStatsByProject(t *testing.T) {
	t.Parallel()

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

	var clientCalls int
	client := &mock.GithubClient{
		StatsByProjectFunc: func(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
			clientCalls++
			return statsResponse, nil
		},
	}
	store := mock.NewKVStore(nil, nil)
	l := logrus.New()

	staleDataClient, err := NewClientWithStaleData(client, store, time.Minute, l)
	if err != nil {
		t.Fatalf("NewClientWithStaleData() error = %v", err)
	}

	stats, err := staleDataClient.StatsByProject(context.Background(), "go", "golang")
	if err != nil {
		t.Fatalf("StatsByProject() subsequent call error = %v", err)
	}
	if !reflect.DeepEqual(stats, statsResponse) {
		t.Fatalf("StatsByProject() returned invalid first project: %+v, want %+v", stats, statsResponse)
	}
	if clientCalls != 1 {
		t.Errorf("StatsByProject() called client %d times, want %d", clientCalls, 1)
	}
}
