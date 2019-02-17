package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/m-zajac/goprojectdemo/app"
	"github.com/m-zajac/goprojectdemo/mock"
	"github.com/pkg/errors"
)

func TestMux(t *testing.T) {
	t.Parallel()

	serviceDelay := time.Millisecond

	tests := []struct {
		name           string
		path           string
		muxTimeout     time.Duration
		wantStatusCode int
	}{
		{
			name:           "valid bestcontributors request",
			path:           "/bestcontributors/go",
			muxTimeout:     time.Second,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "service exceeding handler timeout",
			path:           "/bestcontributors/go",
			muxTimeout:     time.Microsecond,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "invalid path",
			path:           "/invalid_path",
			muxTimeout:     time.Second,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mock.Service{
				MostActiveContributorsFunc: func(
					ctx context.Context,
					language string,
					projectsCount int,
					count int,
				) ([]app.ContributorStats, error) {
					time.Sleep(serviceDelay)

					select {
					case <-ctx.Done():
						return nil, errors.New("context timeout")
					default:
						return nil, nil
					}
				},
			}
			mux := NewMux(service, tt.muxTimeout)

			server := httptest.NewServer(mux)
			defer server.Close()

			url := server.URL + tt.path
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("couldn't connect to server: %v", err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("invalid response status %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}
