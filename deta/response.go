package deta

import (
	"fmt"
	"io"
	"strings"
)

type Response struct {
	StatusCode int
	Data       map[string]interface{}
}

func (r *Response) Ok() bool {
	return r.StatusCode < 300
}

func (r *Response) Error() string {
	if r.StatusCode == 404 {
		key := r.Data["key"].(string)
		return fmt.Sprintf("<Not Found:<<%s>>%d>", key, r.StatusCode)
	}
	if !r.Ok() {
		errors := r.Data["errors"].([]interface{})
		var errs []string
		for _, err := range errors {
			errs = append(errs, err.(string))
		}
		return fmt.Sprintf("<%s: %d>", strings.Join(errs, ","), r.StatusCode)
	}
	return fmt.Sprintf("<Success: %d>", r.StatusCode)
}

type StreamingResponse struct {
	StatusCode int
	Reader     io.ReadCloser
}

func (r *StreamingResponse) Ok() bool {
	return r.StatusCode < 300
}

func (r *StreamingResponse) Error() string {
	return fmt.Sprintf("<%d>", r.StatusCode)
}
