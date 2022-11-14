package deta

import (
	"io"
	"net/http"
)

type service struct {
	key string
	projectId string
}

type httpRequest struct {
	Body io.Reader
	Method string
	Key string
	Path string
}

func (r *httpRequest) do() (*http.Response, error) {
	req, err := http.NewRequest(r.Method, r.Path, r.Body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", r.Key)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}
