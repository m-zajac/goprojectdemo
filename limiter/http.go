package limiter

import (
	"net/http"

	"github.com/pkg/errors"
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
func NewLimitedHTTPDoer(doer HTTPDoer, maxRate int) *LimitedHTTPDoer {
	return &LimitedHTTPDoer{
		doer:    doer,
		limiter: rate.NewLimiter(rate.Limit(maxRate), 1),
	}
}

// Do executes http request. If limit is exceeded, blocks until call rate is within limit.
func (d *LimitedHTTPDoer) Do(r *http.Request) (*http.Response, error) {
	if err := d.limiter.Wait(r.Context()); err != nil {
		return nil, errors.Wrap(err, "waiting for httpDoer limiter")
	}

	return d.doer.Do(r)
}
