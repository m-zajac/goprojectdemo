package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/m-zajac/goprojectdemo/internal/api/http/mock"
	"github.com/m-zajac/goprojectdemo/internal/app"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewContributorsHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		language        string
		setupMock       func(*mock.MockService)
		newRequest      func() *http.Request
		wantStatus      int
		wantBody        string
		wantContentType string
	}{
		{
			name:     "default params values",
			language: "go",
			setupMock: func(m *mock.MockService) {
				m.EXPECT().
					MostActiveContributors(gomock.Any(), "go", defaultHandlerProjectsCountValue, defaultHandlerCountValue).
					Return(nil, nil)
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
			setupMock: func(m *mock.MockService) {
				m.EXPECT().
					MostActiveContributors(gomock.Any(), "go", 3, 7).
					Return(nil, nil)
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
			setupMock: func(m *mock.MockService) {
				m.EXPECT().
					MostActiveContributors(gomock.Any(), "go", defaultHandlerProjectsCountValue, defaultHandlerCountValue).
					Return(nil, app.InvalidRequestError("invalid params"))
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
			setupMock: func(m *mock.MockService) {
				m.EXPECT().
					MostActiveContributors(gomock.Any(), "go", defaultHandlerProjectsCountValue, defaultHandlerCountValue).
					Return(nil, errors.New("error"))
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
			setupMock: func(m *mock.MockService) {
				m.EXPECT().
					MostActiveContributors(gomock.Any(), "go", defaultHandlerProjectsCountValue, defaultHandlerCountValue).
					Return(
						[]app.ContributorStats{
							{
								Commits: 5,
								Contributor: app.Contributor{
									ID:    1,
									Login: "tester",
								},
							},
						},
						nil,
					)
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := mock.NewMockService(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(s)
			}

			l := logrus.New()
			handler := NewContributorsHandler(
				func(*http.Request) string {
					return tt.language
				},
				s,
				l,
			)
			req := tt.newRequest()
			w := httptest.NewRecorder()

			handler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantContentType, w.Header().Get("Content-type"))

			body := w.Body.String()
			body = strings.Trim(body, "\n")
			assert.Equal(t, tt.wantBody, body)
		})
	}
}
