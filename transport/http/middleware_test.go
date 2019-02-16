package http

import (
	"net/http"
	"testing"
	"time"
)

func TestNewTimeoutMiddleware(t *testing.T) {
	t.Parallel()

	m := NewTimeoutMiddleware(time.Millisecond)
	h := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)

		select {
		case <-r.Context().Done():
		default:
			t.Error("request context not canceled")
		}
	}

	r, _ := http.NewRequest(http.MethodGet, "testurl", nil)
	m(h)(nil, r)
}
