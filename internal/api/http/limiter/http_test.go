package limiter

import (
	"context"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/m-zajac/goprojectdemo/internal/mock"
)

func TestLimitedHTTPDoerRate(t *testing.T) {
	maxRate := 500.0
	testTime := 200 * time.Millisecond

	doer := &mock.HTTPDoer{}
	limitedDoer := NewHTTPDoer(doer, maxRate)

	req, _ := http.NewRequest(http.MethodGet, "fakeurl", nil)
	startTime := time.Now()
	var dos int
	for startTime.Add(testTime).After(time.Now()) {
		if _, err := limitedDoer.Do(req); err != nil {
			t.Fatalf("Do() returned error: %v", err)
		}
		dos++
	}

	expectedDos := float64(maxRate) * float64(testTime) / float64(time.Second)
	diff := math.Abs(float64(dos)-expectedDos) / expectedDos
	if diff > 0.1 {
		t.Errorf("unexpected number of Dos: %d, want %d", dos, int(expectedDos))
	}
}

func TestLimitedHTTPDoerTimeout(t *testing.T) {
	doer := &mock.HTTPDoer{}
	limitedDoer := NewHTTPDoer(doer, 1)

	req, _ := http.NewRequest(http.MethodGet, "fakeurl", nil)
	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	if _, err := limitedDoer.Do(req); err != nil {
		t.Fatalf("first Do() returned error: %v", err)
	}

	// Error is expected because of short ctx timeout and low rate limit.
	_, err := limitedDoer.Do(req)
	if err == nil {
		t.Fatal("second Do() didn't return error")
	}
}
