package github

import (
	"reflect"
	"testing"

	"github.com/m-zajac/goprojectdemo/app"
)

func Test_searchResponse_ToProjects(t *testing.T) {
	tests := []struct {
		name     string
		response searchResponse
		want     []app.Project
	}{
		{
			name:     "empty",
			response: searchResponse{},
			want:     []app.Project{},
		},
		{
			name: "2 items",
			response: searchResponse{
				Items: []searchResponseItem{
					{
						ID:   1,
						Name: "x",
						Owner: searchResponseItemOwner{
							Login: "y",
						},
					},
					{
						ID:   2,
						Name: "a",
						Owner: searchResponseItemOwner{
							Login: "b",
						},
					},
				},
			},
			want: []app.Project{
				{
					ID:         1,
					Name:       "x",
					OwnerLogin: "y",
				},
				{
					ID:         2,
					Name:       "a",
					OwnerLogin: "b",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.response.ToProjects(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("searchResponse.ToProjects() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_statsResponse_ToStats(t *testing.T) {
	tests := []struct {
		name     string
		response statsResponse
		want     []app.ContributorStats
	}{
		{
			name:     "empty",
			response: statsResponse{},
			want:     []app.ContributorStats{},
		},
		{
			name: "2 items",
			response: statsResponse{
				{
					Author: statsResponseAuthor{
						ID:    1,
						Login: "x",
					},
					Total: 2,
				},
				{
					Author: statsResponseAuthor{
						ID:    3,
						Login: "y",
					},
					Total: 4,
				},
			},
			want: []app.ContributorStats{
				{
					Commits: 2,
					Contributor: app.Contributor{
						ID:    1,
						Login: "x",
					},
				},
				{
					Commits: 4,
					Contributor: app.Contributor{
						ID:    3,
						Login: "y",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.response.ToStats(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("statsResponse.ToStats() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
