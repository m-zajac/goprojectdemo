package github

import (
	"context"
	"io/ioutil"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/m-zajac/goprojectdemo/internal/adapter/github/mock"
	"github.com/m-zajac/goprojectdemo/internal/app"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var _clientCalls int64
			getClientCalls := func() int {
				v := atomic.LoadInt64(&_clientCalls)
				return int(v)
			}
			clientTokens := make(chan struct{}, 1)

			client := mock.NewMockGithubClient(ctrl)
			client.EXPECT().
				ProjectsByLanguage(gomock.Any(), "go", 1).
				DoAndReturn(func(ctx context.Context, language string, count int) ([]app.Project, error) {
					select {
					case <-clientTokens:
					case <-time.After(time.Second):
						t.Fatal("client locked")
					}

					atomic.AddInt64(&_clientCalls, int64(1))

					return nil, nil
				}).
				AnyTimes()
			client.EXPECT().
				StatsByProject(gomock.Any(), "golang", "go").
				DoAndReturn(func(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
					select {
					case <-clientTokens:
					case <-time.After(time.Second):
						t.Fatal("client locked")
					}

					atomic.AddInt64(&_clientCalls, int64(1))

					return nil, nil
				}).
				AnyTimes()

			storeTokens := make(chan struct{}, 10)
			store := mock.NewKVStore(nil, storeTokens)
			l := logrus.New()
			l.Out = ioutil.Discard

			ttl := time.Minute
			refreshTTL := 10 * time.Second
			staleDataClient, err := NewClientWithStaleData(client, store, ttl, refreshTTL, l)
			require.NoError(t, err)

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

				time.Sleep(100 * time.Millisecond)

				assert.Equal(t, expectedPendingUpdates, pendingUpdates)
				assert.Equal(t, expectedClientCalls, getClientCalls())
				assert.Equal(t, expectedStoreUpdates, store.Updates())
				assert.Equal(t, expectedStoreReads, store.Reads())
			}

			checkNextState("init scheduler")

			// PHASE1: Read with empty db
			t.Log("PHASE1: First call - should read from db, schedule update")
			if err = staleDataClientCall(); !app.IsScheduledForLaterError(err) {
				t.Errorf("phase1: ClientWithStaleData call unexpected error = %v", err)
			}
			expectedStoreReads++
			expectedPendingUpdates++
			checkNextState("phase1: after ClientWithStaleData call")

			t.Log("PHASE1: Next scheduler state - should see empty pending queue, client called and store update")
			expectedPendingUpdates--
			expectedStoreUpdates++
			storeTokens <- struct{}{} // allow store write
			expectedClientCalls++
			clientTokens <- struct{}{} // allow client call
			checkNextState("phase1: after scheduler finishes updates")

			// PHASE2: Read with data already in db
			t.Log("PHASE2: Second call - should read from db but NOT call client")
			if err = staleDataClientCall(); err != nil {
				t.Errorf("phase2: ClientWithStaleData call error = %v", err)
			}
			expectedStoreReads++
			// don't call checkNextState here, nothing is scheduled

			// PHASE3: Read with data in db, but ttl exceeded
			t.Log("PHASE3: Third call - should read from db, schedule update")
			staleDataClient.ttl = 0
			expectedClientCalls++
			clientTokens <- struct{}{} // allow client call
			if err = staleDataClientCall(); !app.IsScheduledForLaterError(err) {
				t.Errorf("phase3: ClientWithStaleData call unexpected error = %v", err)
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	client := mock.NewMockGithubClient(ctrl)
	client.EXPECT().
		ProjectsByLanguage(gomock.Any(), "go", 2).
		Return(projectsResponse, nil)

	store := mock.NewKVStore(nil, nil)
	l := logrus.New()

	staleDataClient, err := NewClientWithStaleData(client, store, time.Minute, time.Minute, l)
	require.NoError(t, err)
	staleDataClient.RunScheduler()

	_, err = staleDataClient.ProjectsByLanguage(context.Background(), "go", 2)
	require.True(t, app.IsScheduledForLaterError(err))

	time.Sleep(10 * time.Millisecond)

	projects, err := staleDataClient.ProjectsByLanguage(context.Background(), "go", 2)
	require.NoError(t, err)
	assert.Equal(t, projectsResponse, projects)
}

func TestClientWithStaleDataStatsByProject(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	client := mock.NewMockGithubClient(ctrl)
	client.EXPECT().
		StatsByProject(gomock.Any(), "go", "golang").
		Return(statsResponse, nil)

	store := mock.NewKVStore(nil, nil)
	l := logrus.New()

	staleDataClient, err := NewClientWithStaleData(client, store, time.Minute, time.Minute, l)
	require.NoError(t, err)
	staleDataClient.RunScheduler()

	_, err = staleDataClient.StatsByProject(context.Background(), "go", "golang")
	require.True(t, app.IsScheduledForLaterError(err))

	time.Sleep(10 * time.Millisecond)

	stats, err := staleDataClient.StatsByProject(context.Background(), "go", "golang")
	require.NoError(t, err)
	assert.Equal(t, statsResponse, stats)
}
