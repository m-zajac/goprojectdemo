package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/m-zajac/goprojectdemo/app"
	"github.com/pkg/errors"
)

// HTTPDoer can execute http request
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client returns details about gihub projects and stats
type Client struct {
	doer           HTTPDoer
	address        string
	authToken      string
	timeout        time.Duration
	acceptWaitTime time.Duration

	projectsResponseMaxSize int
	statsResponseMaxSize    int
}

// NewClient creates new github client.
// authToken is optional.
func NewClient(doer HTTPDoer, address string, authToken string, timeout time.Duration) *Client {
	c := Client{
		doer:           doer,
		address:        address,
		authToken:      authToken,
		timeout:        timeout,
		acceptWaitTime: 2 * time.Second,

		projectsResponseMaxSize: 1024 * 1024 * 10,
		statsResponseMaxSize:    1024 * 1024 * 30,
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
		return nil, errors.Wrap(err, "invalid url")
	}

	v := make(url.Values)
	v.Set("q", "language:"+language)
	v.Set("sort", "stars")
	v.Set("per_page", strconv.Itoa(count))
	u.RawQuery = v.Encode()

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating http request")
	}

	body, _, err := c.makeRequest(ctx, httpReq, c.projectsResponseMaxSize)
	if err != nil {
		return nil, err
	}

	var resp searchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, errors.Wrap(err, "unmarshalling response")
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
		return nil, errors.Wrap(err, "invalid url")
	}

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating http request")
	}

	// Github returns status 202 when processing data.
	// Should wait a bit and try again
	var tries int
	var body []byte
	for {
		tries++
		b, code, err := c.makeRequest(ctx, httpReq, 1024*1024*100)
		if err != nil {
			return nil, err
		}
		if code == http.StatusAccepted {
			if tries < 5 {
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
		return nil, errors.Wrap(err, "unmarshalling response")
	}

	return resp.ToStats(), nil
}

func (c *Client) makeRequest(ctx context.Context, req *http.Request, maxBytes int) ([]byte, int, error) {
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "token "+c.authToken)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.doer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, 0, errors.Wrap(err, "doing http request")
	}
	// Always drain body before close to allow connection reuse
	// See: http://tleyden.github.io/blog/2016/11/21/tuning-the-go-http-client-library-for-load-testing/
	defer func() {
		io.CopyN(ioutil.Discard, resp.Body, 1024)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNoContent {
		return nil, resp.StatusCode, nil
	}
	if resp.StatusCode/100 > 3 {
		return nil, resp.StatusCode, errors.Errorf("got invalid http status code: %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(io.LimitReader(resp.Body, int64(maxBytes)))
	if err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "reading http response body")
	}

	return b, resp.StatusCode, nil
}
