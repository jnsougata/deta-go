package deta

import (
	"fmt"
	"io"
	"net/http"
)

const baseRoot = "https://database.deta.sh/v1"

type base struct {
	Name    string
	service *service
}

func (b *base) Put(items ...map[string]interface{}) ([]http.Response, []error) {
	var bodies [][]map[string]interface{}
	if len(items) > 25 {
		bodies = autoSlice(items, 25)
	} else {
		bodies = append(bodies, items)
	}
	respChannel := make(chan http.Response, len(bodies))
	errChannel := make(chan error, len(bodies))
	for _, body := range bodies {
		go func(body []map[string]interface{}) {
			data := map[string]interface{}{
				"items": body,
			}
			reader, err := interfaceReader(data)
			if err != nil {
				errChannel <- err
				return
			}
			req := httpRequest{
				Body:   reader,
				Method: "PUT",
				Path:   fmt.Sprintf("%s/%s/%s/items", baseRoot, b.service.projectId, b.Name),
				Key:    b.service.key,
			}
			resp, err := req.do()
			if err != nil {
				errChannel <- err
			} else {
				errChannel <- nil
			}
			respChannel <- *resp
		}(body)
	}
	responses := make([]http.Response, len(bodies))
	errors := make([]error, len(bodies))
	for i := 0; i < len(bodies); i++ {
		responses[i] = <-respChannel
		errors[i] = <-errChannel
	}
	return responses, errors
}

func (b *base) fetch(last string) (*http.Response, error) {
	var body io.Reader
	path := fmt.Sprintf("%s/%s/%s/query", baseRoot, b.service.projectId, b.Name)
	if last != "" {
		reader, _ := interfaceReader(map[string]interface{}{"last": last, "query": []map[string]interface{}{}})
		body = reader
	} else {
		reader, _ := interfaceReader(map[string]interface{}{"query": []map[string]interface{}{}})
		body = reader
	}
	req := httpRequest{
		Body:   body,
		Method: "POST",
		Path:   path,
		Key:    b.service.key,
	}
	return req.do()
}

func (b *base) Get(keys ...string) ([]map[string]interface{}, error) {
	var container []map[string]interface{}
	if len(keys) == 0 {
		var last string
		resp, err := b.fetch(last)
		if err != nil {
			return nil, err
		}
		data, err := responseReader(resp)
		if err != nil {
			return nil, err
		}
		items := data["items"]
		for _, item := range items.([]interface{}) {
			container = append(container, item.(map[string]interface{}))
		}
		lastValue, ok := data["paging"].(map[string]interface{})["last"]
		if ok {
			last = lastValue.(string)
			for {
				resp, err := b.fetch(last)
				if err != nil {
					return nil, err
				}
				data, err := responseReader(resp)
				if err != nil {
					return nil, err
				}
				items := data["items"]
				for _, item := range items.([]interface{}) {
					container = append(container, item.(map[string]interface{}))
				}
				lastValue, ok := data["paging"].(map[string]interface{})["last"]
				if ok {
					last = lastValue.(string)
				} else {
					break
				}
			}
		}
	} else if len(keys) == 1 {
		req := httpRequest{
			Body:   nil,
			Method: "GET",
			Path:   fmt.Sprintf("%s/%s/%s/items/%s", baseRoot, b.service.projectId, b.Name, keys[0]),
			Key:    b.service.key,
		}
		resp, err := req.do()
		if err != nil {
			return nil, err
		}
		data, err := responseReader(resp)
		if err != nil {
			return nil, err
		}
		container = append(container, data)
	} else {
		responses := make(chan map[string]interface{}, len(keys))
		for _, key := range keys {
			go func(key string) {
				req := httpRequest{
					Body:   nil,
					Method: "GET",
					Path:   fmt.Sprintf("%s/%s/%s/items/%s", baseRoot, b.service.projectId, b.Name, key),
					Key:    b.service.key,
				}
				resp, err := req.do()
				if err != nil {
					responses <- nil
				}
				data, err := responseReader(resp)
				if err != nil {
					responses <- map[string]interface{}{}
				}
				responses <- data
			}(key)
		}
		for i := 0; i < len(keys); i++ {
			container = append(container, <-responses)
		}
	}
	return container, nil
}

func (b *base) Delete(keys ...string) []map[string]interface{} {
	respChannel := make(chan map[string]interface{}, len(keys))
	for _, key := range keys {
		go func(key string) {
			req := httpRequest{
				Body:   nil,
				Method: "DELETE",
				Path:   fmt.Sprintf("%s/%s/%s/items/%s", baseRoot, b.service.projectId, b.Name, key),
				Key:    b.service.key,
			}
			resp, _ := req.do()
			data, _ := responseReader(resp)
			respChannel <- data
		}(key)
	}
	responses := make([]map[string]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		responses[i] = <-respChannel
	}
	return responses
}

func (b *base) Insert(key string , item map[string]interface{}) (map[string]interface{}, error) {
	if key != "" {
		item["key"] = key
	}
	reader, _ := interfaceReader(map[string]interface{}{"item": item})
	fmt.Println(item)
	req := httpRequest{
		Body:   reader,
		Method: "POST",
		Path:   fmt.Sprintf("%s/%s/%s/items", baseRoot, b.service.projectId, b.Name),
		Key:    b.service.key,
	}
	resp, err := req.do()
	if err != nil {
		return nil, err
	}
	return responseReader(resp)
}

func (b *base) Update(key string) *updater {
	return &updater{
		key:  key,
		baseName: b.Name,
		service: b.service,
		updates: make(map[string]interface{}),
	}
}

func (b *base) Fetch(query *query) map[string]interface{} {
	body := []map[string]interface{}{}
	if len(query.ors) > 0 {
		body = query.ors
	} else if len(query.values) > 0 {
		body = append(body, query.values)
	} else {
		body = []map[string]interface{}{}
	}
	queryBody := map[string]interface{}{"query": body}
	if query.Limit != 0 {
		queryBody["limit"] = query.Limit
	}
	if query.Last != "" {
		queryBody["last"] = query.Last
	}
	reader, _ := interfaceReader(queryBody)
	req := httpRequest{
		Body:   reader,
		Method: "POST",
		Path:   fmt.Sprintf("%s/%s/%s/query", baseRoot, b.service.projectId, b.Name),
		Key:    b.service.key,
	}
	resp, _ := req.do()
	data, _ := responseReader(resp)
	return data
}
