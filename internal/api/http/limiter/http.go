package limiter

import (
	"fmt"
	"net/http"

	"github.com/m-zajac/goprojectdemo/internal/app"
	"golang.org/x/time/rate"
)

// HTTPDoer can execute http request.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// limitedHTTPDoer wraps HTTPDoer and allows Dos with maximum rate limit.
type limitedHTTPDoer struct {
	doer    HTTPDoer
	limiter *rate.Limiter
}

// NewHTTPDoer creates LimitedHTTPDoer instance.
// maxRate - maximum number of Dos per second.
func NewHTTPDoer(doer HTTPDoer, maxRate float64) HTTPDoer {
	return &limitedHTTPDoer{
		doer:    doer,
		limiter: rate.NewLimiter(rate.Limit(maxRate), 1),
	}
}

// Do executes http request. If limit is exceeded, blocks until call rate is within limit.
func (d *limitedHTTPDoer) Do(r *http.Request) (*http.Response, error) {
	if err := d.limiter.Wait(r.Context()); err != nil {
		return nil, app.TooManyRequestsError(fmt.Sprintf("waiting for httpDoer limiter: %v", err))
	}

	return d.doer.Do(r)
}
