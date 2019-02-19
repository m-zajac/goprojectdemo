package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/m-zajac/goprojectdemo/app"
	"github.com/sirupsen/logrus"
)

// Service can return most active contributors.
type Service interface {
	MostActiveContributors(
		ctx context.Context,
		language string,
		projectsCount int,
		count int,
	) ([]app.ContributorStats, error)
}

// NewMux creates router for app's http server.
func NewMux(service Service, timeout time.Duration, l logrus.FieldLogger) *http.ServeMux {
	timeoutMiddleware := NewTimeoutMiddleware(timeout)

	contributorsPath := "/bestcontributors/"
	contributorsHandler := NewContributorsHandler(
		func(r *http.Request) string {
			return strings.TrimPrefix(r.URL.Path, contributorsPath)
		},
		service,
		l.WithField("handler", "contributorsHandler"),
	)
	contributorsHandler = timeoutMiddleware(contributorsHandler)

	m := http.NewServeMux()
	m.HandleFunc(contributorsPath, contributorsHandler)

	return m
}
