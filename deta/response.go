package deta

type Response struct {
	StatusCode int
	Data       map[string]interface{}
}

func (r *Response) Ok() bool {
	return r.StatusCode < 300
}

func (r *Response) Issue() string {
	return ""
}
