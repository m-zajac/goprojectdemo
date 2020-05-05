package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/m-zajac/goprojectdemo/internal/app"
)

// HTTPDoer can execute http request.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client returns details about gihub projects and stats.
// This struct is an adapter for app.GithubClient.
//go:generate mockgen -destination mock/githubcli.go -package mock github.com/m-zajac/goprojectdemo/internal/app GithubClient
type Client struct {
	doer           HTTPDoer
	address        string
	authToken      string
	acceptWaitTime time.Duration

	projectsResponseMaxSize int
	statsResponseMaxSize    int
	numRetriesOnAccepted    int
}

var _ app.GithubClient = &Client{}

// NewClient creates new github client.
// authToken is optional.
func NewClient(doer HTTPDoer, address string, authToken string) *Client {
	c := Client{
		doer:           doer,
		address:        address,
		authToken:      authToken,
		acceptWaitTime: 5 * time.Second,

		projectsResponseMaxSize: 1024 * 1024 * 10,
		statsResponseMaxSize:    1024 * 1024 * 30,
		numRetriesOnAccepted:    7,
	}

	return &c
}

// ProjectsByLanguage returns projects by given programming language name.
func (c *Client) ProjectsByLanguage(ctx context.Context, language string, count int) ([]app.Project, error) {
	if language == "" {
		return nil, app.InvalidRequestError("lanuage cannot be empty")
	}
	if count < 1 || count > 99 {
		return nil, app.InvalidRequestError("count must be in range <1..99>")
	}

	u, err := url.Parse(c.address + "/search/repositories")
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	v := make(url.Values)
	v.Set("q", "language:"+language)
	v.Set("sort", "stars")
	v.Set("per_page", strconv.Itoa(count))
	u.RawQuery = v.Encode()

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	body, _, err := c.makeRequest(ctx, httpReq, c.projectsResponseMaxSize)
	if err != nil {
		return nil, fmt.Errorf("making http request: %w", err)
	}

	var resp searchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshalling response: %w", err)
	}

	return resp.ToProjects(), nil
}

// StatsByProject returns stats by given github project params.
func (c *Client) StatsByProject(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
	if name == "" {
		return nil, app.InvalidRequestError("project's name cannot be empty")
	}
	if owner == "" {
		return nil, app.InvalidRequestError("project's owner login cannot be empty")
	}

	u, err := url.Parse(c.address + fmt.Sprintf("/repos/%s/%s/stats/contributors", owner, name))
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	// Github returns status 202 when processing data.
	// Should wait a bit and try again.
	var tries int
	var body []byte
	for {
		tries++
		b, code, err := c.makeRequest(ctx, httpReq, 1024*1024*100)
		if err != nil {
			return nil, fmt.Errorf("making http request: %w", err)
		}
		if code == http.StatusAccepted {
			if tries < c.numRetriesOnAccepted {
				time.Sleep(c.acceptWaitTime)
				continue
			}
			return nil, errors.New("too many reties with status 202")
		}
		body = b
		break
	}

	var resp statsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshalling response: %w", err)
	}

	return resp.ToStats(), nil
}

func (c *Client) makeRequest(ctx context.Context, req *http.Request, maxBytes int) ([]byte, int, error) {
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "token "+c.authToken)
	}

	resp, err := c.doer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, 0, fmt.Errorf("doing http request: %w", err)
	}
	// Always drain body before close to allow connection reuse.
	// See: http://tleyden.github.io/blog/2016/11/21/tuning-the-go-http-client-library-for-load-testing/
	defer func() {
		_, _ = io.CopyN(ioutil.Discard, resp.Body, 1024)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNoContent {
		return nil, resp.StatusCode, nil
	}
	if resp.StatusCode/100 > 3 {
		if c.checkRateLimitExceeded(&resp.Header) {
			return nil, resp.StatusCode, errors.New("rate limit exceeded")
		}
		return nil, resp.StatusCode, fmt.Errorf("got invalid http status code: %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(io.LimitReader(resp.Body, int64(maxBytes)))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading http response body: %w", err)
	}

	return b, resp.StatusCode, nil
}

func (c *Client) checkRateLimitExceeded(h *http.Header) bool {
	if s := h.Get("X-RateLimit-Remaining"); s != "" {
		if limit, err := strconv.Atoi(s); err == nil && limit == 0 {
			return true
		}
	}
	return false
}
