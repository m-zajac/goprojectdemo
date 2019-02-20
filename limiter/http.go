package limiter

import (
	"fmt"
	"net/http"

	"github.com/m-zajac/goprojectdemo/app"
	"golang.org/x/time/rate"
)

// HTTPDoer can execute http request.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// LimitedHTTPDoer wraps HTTPDoer and allows Dos with maximum rate limit.
type LimitedHTTPDoer struct {
	doer    HTTPDoer
	limiter *rate.Limiter
}

// NewLimitedHTTPDoer creates LimitedHTTPDoer instance.
// maxRate - maximum number of Dos per second.
func NewLimitedHTTPDoer(doer HTTPDoer, maxRate float64) *LimitedHTTPDoer {
	return &LimitedHTTPDoer{
		doer:    doer,
		limiter: rate.NewLimiter(rate.Limit(maxRate), 1),
	}
}

// Do executes http request. If limit is exceeded, blocks until call rate is within limit.
func (d *LimitedHTTPDoer) Do(r *http.Request) (*http.Response, error) {
	if err := d.limiter.Wait(r.Context()); err != nil {
		return nil, app.TooManyRequestsError(fmt.Sprintf("waiting for httpDoer limiter: %v", err))
	}

	return d.doer.Do(r)
}
