package github

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/m-zajac/goprojectdemo/internal/app"
	"github.com/m-zajac/goprojectdemo/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ProjectsByLanguage(t *testing.T) {
	t.Parallel()

	var bigDataBlob []byte
	for i := 0; i < 1024*1024*100; i++ {
		bigDataBlob = append(bigDataBlob, 'x')
	}

	tests := []struct {
		name     string
		doer     *mock.HTTPDoer
		language string
		count    int
		want     []app.Project
		wantErr  bool
	}{
		{
			name:     "empty language",
			language: "",
			count:    1,
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid count",
			language: "go",
			count:    -1,
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid count",
			language: "go",
			count:    111,
			want:     nil,
			wantErr:  true,
		},
		{
			name: "status ok, body ok",
			doer: &mock.HTTPDoer{
				Statuses: []int{http.StatusOK},
				Bodies: [][]byte{
					[]byte(`{
						"total_count": 409167,
						"incomplete_results": false,
						"items": [
							{
								"id": 23096959,
								"node_id": "MDEwOlJlcG9zaXRvcnkyMzA5Njk1OQ==",
								"name": "go",
								"full_name": "golang/go",
								"owner": {
									"login": "golang",
									"id": 4314092
								},
								"language": "Go"
							}
						]
					}`),
				},
			},
			language: "go",
			count:    1,
			want: []app.Project{
				{
					ID:         23096959,
					Name:       "go",
					OwnerLogin: "golang",
				},
			},
			wantErr: false,
		},
		{
			name: "status not ok",
			doer: &mock.HTTPDoer{
				Statuses: []int{http.StatusInternalServerError},
			},
			language: "go",
			count:    1,
			want:     nil,
			wantErr:  true,
		},
		{
			name: "status ok, body unexpectedly large",
			doer: &mock.HTTPDoer{
				Statuses: []int{http.StatusOK},
				Bodies: [][]byte{
					bigDataBlob,
				},
			},
			language: "go",
			count:    1,
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(tt.doer, "https://fake", "token")
			got, err := c.ProjectsByLanguage(
				context.Background(),
				tt.language,
				tt.count,
			)
			require.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)

			if tt.doer == nil {
				return
			}

			require.Len(t, tt.doer.Responses, 1)
			req := tt.doer.Responses[0].Request
			assert.Equal(t, "language:"+tt.language, req.URL.Query().Get("q"))
			assert.Equal(t, "stars", req.URL.Query().Get("sort"))
			assert.Equal(t, strconv.Itoa(tt.count), req.URL.Query().Get("per_page"))

			checkAPIHeaders(req, t)
		})
	}
}

func TestClient_StatsByProject(t *testing.T) {
	t.Parallel()

	var bigDataBlob []byte
	for i := 0; i < 1024*1024*100; i++ {
		bigDataBlob = append(bigDataBlob, 'x')
	}

	validStatsJSON := []byte(`[
		{
			"total": 3,
			"weeks": [
				{
					"w": 1530403200,
					"a": 0,
					"d": 0,
					"c": 1
				},
				{
					"w": 1531008000,
					"a": 0,
					"d": 0,
					"c": 2
				}
			],
			"author": {
				"login": "minderov",
				"id": 15854038
			}
		},
		{
			"total": 7,
			"weeks": [
				{
					"w": 1530403200,
					"a": 0,
					"d": 0,
					"c": 3
				},
				{
					"w": 1531008000,
					"a": 0,
					"d": 0,
					"c": 4
				}
			],
			"author": {
				"login": "KarandikarMihir",
				"id": 17466938
			}
		}
	]`)

	tests := []struct {
		name         string
		doer         *mock.HTTPDoer
		projectName  string
		owner        string
		want         []app.ContributorStats
		wantErr      bool
		wantAPICalls int
	}{
		{
			name:         "empty owner",
			projectName:  "100-Days-Of-ML-Code",
			owner:        "",
			want:         nil,
			wantErr:      true,
			wantAPICalls: 0,
		},
		{
			name:         "empty project name",
			projectName:  "",
			owner:        "Avik-Jain",
			want:         nil,
			wantErr:      true,
			wantAPICalls: 0,
		},
		{
			name: "status ok, body ok",
			doer: &mock.HTTPDoer{
				Statuses: []int{http.StatusOK},
				Bodies: [][]byte{
					validStatsJSON,
				},
			},
			projectName: "100-Days-Of-ML-Code",
			owner:       "Avik-Jain",
			want: []app.ContributorStats{
				{
					Commits: 3,
					Contributor: app.Contributor{
						ID:    15854038,
						Login: "minderov",
					},
				},
				{
					Commits: 7,
					Contributor: app.Contributor{
						ID:    17466938,
						Login: "KarandikarMihir",
					},
				},
			},
			wantErr:      false,
			wantAPICalls: 1,
		},
		{
			name: "status not ok",
			doer: &mock.HTTPDoer{
				Statuses: []int{http.StatusInternalServerError},
			},
			projectName:  "100-Days-Of-ML-Code",
			owner:        "Avik-Jain",
			want:         nil,
			wantErr:      true,
			wantAPICalls: 1,
		},
		{
			name: "status ok, body unexpectedly large",
			doer: &mock.HTTPDoer{
				Statuses: []int{http.StatusOK},
				Bodies: [][]byte{
					bigDataBlob,
				},
			},
			projectName:  "100-Days-Of-ML-Code",
			owner:        "Avik-Jain",
			want:         nil,
			wantErr:      true,
			wantAPICalls: 1,
		},
		{
			name: "2 time 202, then valid response",
			doer: &mock.HTTPDoer{
				Statuses: []int{
					http.StatusAccepted,
					http.StatusAccepted,
					http.StatusOK,
				},
				Bodies: [][]byte{
					{},
					{},
					validStatsJSON,
				},
			},
			projectName: "100-Days-Of-ML-Code",
			owner:       "Avik-Jain",
			want: []app.ContributorStats{
				{
					Commits: 3,
					Contributor: app.Contributor{
						ID:    15854038,
						Login: "minderov",
					},
				},
				{
					Commits: 7,
					Contributor: app.Contributor{
						ID:    17466938,
						Login: "KarandikarMihir",
					},
				},
			},
			wantErr:      false,
			wantAPICalls: 3,
		},
		{
			name: "got 202 too many times",
			doer: &mock.HTTPDoer{
				Statuses: []int{
					http.StatusAccepted,
					http.StatusAccepted,
					http.StatusAccepted,
					http.StatusAccepted,
					http.StatusAccepted,
					http.StatusAccepted,
					http.StatusAccepted,
					http.StatusOK,
				},
				Bodies: [][]byte{
					{},
					{},
					{},
					{},
					{},
					{},
					{},
					validStatsJSON,
				},
			},
			projectName:  "100-Days-Of-ML-Code",
			owner:        "Avik-Jain",
			want:         nil,
			wantErr:      true,
			wantAPICalls: 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(tt.doer, "https://fake", "token")
			c.acceptWaitTime = 100 * time.Millisecond
			got, err := c.StatsByProject(
				context.Background(),
				tt.projectName,
				tt.owner,
			)
			require.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)

			if tt.doer == nil {
				return
			}

			require.Equal(t, tt.wantAPICalls, len(tt.doer.Responses))

			req := tt.doer.Responses[0].Request
			checkAPIHeaders(req, t)
		})
	}
}

func checkAPIHeaders(r *http.Request, t *testing.T) {
	assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
	assert.Contains(t, r.Header.Get("Authorization"), "token ")
}
