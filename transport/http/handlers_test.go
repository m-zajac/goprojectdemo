package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/m-zajac/goprojectdemo/app"
	"github.com/m-zajac/goprojectdemo/mock"
	"github.com/pkg/errors"
)

func TestNewContributorsHandler(t *testing.T) {
	t.Parallel()

	defaultProjectsCount := 5
	defaultCount := 10

	tests := []struct {
		name            string
		language        string
		newService      func(*testing.T) *mock.Service
		newRequest      func() *http.Request
		wantStatus      int
		wantBody        string
		wantContentType string
	}{
		{
			name:     "default params values",
			language: "go",
			newService: func(t *testing.T) *mock.Service {
				return &mock.Service{
					MostActiveContributorsFunc: func(
						ctx context.Context,
						language string,
						projectsCount int,
						count int,
					) ([]app.ContributorStats, error) {
						if projectsCount != defaultProjectsCount {
							t.Errorf("service: invalid projectsCount %d, want %d", projectsCount, defaultProjectsCount)
						}
						if count != defaultCount {
							t.Errorf("service: invalid count %d, want %d", count, defaultCount)
						}
						return nil, nil
					},
				}
			},
			newRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "testurl", nil)
				return r
			},
			wantStatus:      http.StatusOK,
			wantBody:        `{"language":"go","contributors":[]}`,
			wantContentType: "application/json; charset=utf-8",
		},
		{
			name:     "params values from url query",
			language: "go",
			newService: func(t *testing.T) *mock.Service {
				return &mock.Service{
					MostActiveContributorsFunc: func(
						ctx context.Context,
						language string,
						projectsCount int,
						count int,
					) ([]app.ContributorStats, error) {
						wantProjectsCount := 3
						wantCount := 7
						if projectsCount != wantProjectsCount {
							t.Errorf("service: invalid projectsCount %d, want %d", projectsCount, wantProjectsCount)
						}
						if count != wantCount {
							t.Errorf("service: invalid count %d, want %d", count, wantCount)
						}
						return nil, nil
					},
				}
			},
			newRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "testurl?count=7&projectsCount=3", nil)
				return r
			},
			wantStatus:      http.StatusOK,
			wantBody:        `{"language":"go","contributors":[]}`,
			wantContentType: "application/json; charset=utf-8",
		},
		{
			name:     "bad request",
			language: "go",
			newService: func(*testing.T) *mock.Service {
				return &mock.Service{
					MostActiveContributorsFunc: func(
						ctx context.Context,
						language string,
						projectsCount int,
						count int,
					) ([]app.ContributorStats, error) {
						return nil, app.InvalidRequestError("invalid params")
					},
				}
			},
			newRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "testurl", nil)
				return r
			},
			wantStatus:      http.StatusBadRequest,
			wantBody:        `invalid params`,
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			name:     "service error",
			language: "go",
			newService: func(*testing.T) *mock.Service {
				return &mock.Service{
					MostActiveContributorsFunc: func(
						ctx context.Context,
						language string,
						projectsCount int,
						count int,
					) ([]app.ContributorStats, error) {
						return nil, errors.New("error")
					},
				}
			},
			newRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "testurl", nil)
				return r
			},
			wantStatus:      http.StatusInternalServerError,
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			name:     "valid response",
			language: "go",
			newService: func(*testing.T) *mock.Service {
				return &mock.Service{
					MostActiveContributorsFunc: func(
						ctx context.Context,
						language string,
						projectsCount int,
						count int,
					) ([]app.ContributorStats, error) {
						return []app.ContributorStats{
							{
								Commits: 5,
								Contributor: app.Contributor{
									ID:    1,
									Login: "tester",
								},
							},
						}, nil
					},
				}
			},
			newRequest: func() *http.Request {
				r, _ := http.NewRequest(http.MethodGet, "testurl", nil)
				return r
			},
			wantStatus:      http.StatusOK,
			wantBody:        `{"language":"go","contributors":[{"name":"tester","commits":5}]}`,
			wantContentType: "application/json; charset=utf-8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewContributorsHandler(
				func(*http.Request) string {
					return tt.language
				},
				tt.newService(t),
			)
			req := tt.newRequest()
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("invalid handler response status %d, want %d", w.Code, tt.wantStatus)
			}
			body := w.Body.String()
			body = strings.Trim(body, "\n")
			if body != tt.wantBody {
				t.Errorf("invalid body\n\tgot:\n%s\n\twant:\n%s", body, tt.wantBody)
			}

			contentType := w.Header().Get("Content-type")
			if contentType != tt.wantContentType {
				t.Errorf("invalid content type '%s', want '%s'", contentType, tt.wantContentType)
			}
		})
	}
}
