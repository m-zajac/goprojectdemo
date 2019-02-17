package http

import (
	"net/http"
	"strconv"

	jsoniter "github.com/json-iterator/go"

	"github.com/m-zajac/goprojectdemo/app"
)

type contributor struct {
	Name    string `json:"name"`
	Commits int    `json:"commits"`
}

type contributorsResponse struct {
	Language     string        `json:"language"`
	Contributors []contributor `json:"contributors"`
}

func newContributorsResponse(language string, contributions []app.ContributorStats) contributorsResponse {
	contributors := make([]contributor, 0, len(contributions))
	for _, c := range contributions {
		contributors = append(contributors, contributor{
			Name:    c.Contributor.Login,
			Commits: c.Commits,
		})
	}

	return contributorsResponse{
		Language:     language,
		Contributors: contributors,
	}
}

// NewContributorsHandler creates handlerfunc returning contributions response.
func NewContributorsHandler(
	getLanguage func(*http.Request) string,
	service Service,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := getLanguage(r)
		count := getIntParam(r, "count", 10)
		projectsCount := getIntParam(r, "projectsCount", 5)

		contributions, err := service.MostActiveContributors(r.Context(), lang, projectsCount, count)
		if err != nil {
			if app.IsInvalidRequestError(err) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		response := newContributorsResponse(lang, contributions)

		w.Header().Set("Content-type", "application/json; charset=utf-8")
		_ = jsoniter.ConfigFastest.NewEncoder(w).Encode(response)
	}
}

func getIntParam(r *http.Request, name string, defaultValue int) int {
	value := defaultValue
	if vs := r.URL.Query().Get(name); vs != "" {
		if v, err := strconv.Atoi(vs); err == nil && v > 0 && v < 100 {
			value = v
		}
	}

	return value
}
