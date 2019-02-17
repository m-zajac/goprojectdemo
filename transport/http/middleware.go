package http

import (
	"context"
	"net/http"
	"time"
)

// NewTimeoutMiddleware creates middleware that cancels requests context after given time.
func NewTimeoutMiddleware(timeout time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			h(w, r)
		}
	}
}
