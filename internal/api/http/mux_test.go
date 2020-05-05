package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/m-zajac/goprojectdemo/internal/api/http/mock"
	"github.com/m-zajac/goprojectdemo/internal/app"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mock.NewMockService(ctrl)
			service.EXPECT().
				MostActiveContributors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(
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
				}).
				MaxTimes(1)

			l := logrus.New()
			mux := NewMux(service, tt.muxTimeout, l)

			server := httptest.NewServer(mux)
			defer server.Close()

			url := server.URL + tt.path
			resp, err := http.Get(url)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}
