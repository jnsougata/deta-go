package deta

import "fmt"

type query struct {
	Last string
	Limit int
	values map[string]interface{}
	ors    []map[string]interface{}
}

func (q *query) Equals(attrs map[string]interface{}) {
	for k, v := range attrs {
		q.values[k] = v
	}
}

func (q *query) NotEquals(attrs map[string]interface{}) {
	for k, v := range attrs {
		q.values[fmt.Sprintf("%s?ne", k)] = v
	}
}

func (q *query) GreaterThan(attrs map[string]interface{}) {
	for k, v := range attrs {
		q.values[fmt.Sprintf("%s?gt", k)] = v
	}
}

func (q *query) GreaterThanOrEqual(attrs map[string]interface{}) {
	for k, v := range attrs {
		q.values[fmt.Sprintf("%s?gte", k)] = v
	}
}

func (q *query) LessThan(attrs map[string]interface{}) {
	for k, v := range attrs {
		q.values[fmt.Sprintf("%s?lt", k)] = v
	}
}

func (q *query) LessThanOrEqual(attrs map[string]interface{}) {
	for k, v := range attrs {
		q.values[fmt.Sprintf("%s?lte", k)] = v
	}
}

func (q *query) Prefix(attrs map[string]interface{}) {
	for k, v := range attrs {
		q.values[fmt.Sprintf("%s?pfx", k)] = v
	}
}

func (q *query) Range(attrs map[string]interface{}) {
	for k, v := range attrs {
		q.values[fmt.Sprintf("%s?r", k)] = v
	}
}

func (q *query) Contains(attrs map[string]interface{}) {
	for k, v := range attrs {
		q.values[fmt.Sprintf("%s?contains", k)] = v
	}
}

func (q *query) NotContains(attrs map[string]interface{}) {
	for k, v := range attrs {
		q.values[fmt.Sprintf("%s?not_contains", k)] = v
	}
}

func (q *query) Or(queries ...query) *query {
	for _, query := range queries {
		q.ors = append(q.ors, query.values)
	}
	return q
}

func NewQuery(last string, limit int) *query {
	return &query{
		values: map[string]interface{}{},
		ors:    []map[string]interface{}{},
		Last:   last,
		Limit:  limit,
	}
}