package github

import (
	"context"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/m-zajac/goprojectdemo/app"
	"github.com/m-zajac/goprojectdemo/mock"
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
		timeout  time.Duration
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
			timeout := tt.timeout
			if timeout == 0 {
				timeout = time.Minute
			}
			c := NewClient(tt.doer, "https://fake", "token")
			got, err := c.ProjectsByLanguage(
				context.Background(),
				tt.language,
				tt.count,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ProjectsByLanguage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.ProjectsByLanguage() = %+v, want %+v", got, tt.want)
			}

			if tt.doer == nil {
				return
			}

			if v := len(tt.doer.Responses); v != 1 {
				t.Fatalf("Client.ProjectsByLanguage() called api %d times, should be just one", v)
			}

			req := tt.doer.Responses[0].Request
			if v := req.URL.Query().Get("q"); v != "language:"+tt.language {
				t.Errorf("Client.ProjectsByLanguage() has invalid q query param: %s", v)
			}
			if v := req.URL.Query().Get("sort"); v != "stars" {
				t.Errorf("Client.ProjectsByLanguage() has invalid sort query param: %s", v)
			}
			if v := req.URL.Query().Get("per_page"); v != strconv.Itoa(tt.count) {
				t.Errorf("Client.ProjectsByLanguage() has invalid per_page query param: %s", v)
			}

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
		timeout      time.Duration
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.StatsByProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.StatsByProject() = %+v, want %+v", got, tt.want)
			}

			if tt.doer == nil {
				return
			}

			if v := len(tt.doer.Responses); v != tt.wantAPICalls {
				t.Fatalf("Client.StatsByProject() called api %d times, want %d", v, tt.wantAPICalls)
			}

			req := tt.doer.Responses[0].Request
			checkAPIHeaders(req, t)
		})
	}
}

func checkAPIHeaders(r *http.Request, t *testing.T) {
	if v := r.Header.Get("Accept"); v != "application/vnd.github.v3+json" {
		t.Errorf("Api called with invalid header Accept: %s", v)
	}
	if v := r.Header.Get("Authorization"); !strings.HasPrefix(v, "token ") {
		t.Errorf("Api called with invalid header Authorization: %s", v)
	}
}
