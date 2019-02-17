package mock

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// HTTPDoer mocks http.Client.
type HTTPDoer struct {
	Statuses []int
	Bodies   [][]byte
	Headers  []http.Header

	DoFunc    func(*http.Request) (*http.Response, error)
	Responses []*http.Response

	i int
}

// Do fakes executing http request.
func (d *HTTPDoer) Do(r *http.Request) (*http.Response, error) {
	defer func() {
		d.i++
	}()

	if d.DoFunc != nil {
		return d.DoFunc(r)
	}

	status := http.StatusOK
	if len(d.Statuses) > 0 {
		status = d.Statuses[d.i%len(d.Statuses)]
	}
	var data []byte
	if len(d.Bodies) > 0 {
		data = d.Bodies[d.i%len(d.Bodies)]
	}
	buf := bytes.NewBuffer(data)
	body := ioutil.NopCloser(buf)

	header := http.Header{}
	if len(d.Headers) > 0 {
		header = d.Headers[d.i%len(d.Bodies)]
	}

	response := &http.Response{
		StatusCode: status,
		Body:       body,
		Header:     header,
		Request:    r,
	}
	d.Responses = append(d.Responses, response)

	return response, nil
}
