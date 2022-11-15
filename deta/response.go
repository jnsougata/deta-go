package deta

type response struct {
	StatusCode int
	Data       map[string]interface{}
}

func (r *response) Ok() bool {
	return r.StatusCode < 300
}

func (r *response) Issue() string {
	return ""
}
