package deta

import (
	"fmt"
	"io"
	"net/http"
)

const baseHost = "https://database.deta.sh/v1"

type base struct {
	Name    string
	service *service
}

// Put upserts a new item into the collection with the given key.
// If the key is not provided, a new key will be generated.
// A Put can upsert 25 items at a time. If more than 25 items are provided,
// it will automatically slice the items into chunks of 25 and make multiple requests.
func (b *base) Put(items ...map[string]interface{}) []*Response {
	var chunks [][]map[string]interface{}
	if len(items) > 25 {
		chunks = autoSlice(items, 25)
	} else {
		chunks = append(chunks, items)
	}
	respChannel := make(chan *Response, len(chunks))
	errChannel := make(chan error, len(chunks))
	for _, body := range chunks {
		go func(body []map[string]interface{}) {
			d := map[string]interface{}{"items": body}
			reader, err := interfaceReader(d)
			if err != nil {
				errChannel <- err
				return
			}
			req := httpRequest{
				Body:   reader,
				Method: "PUT",
				Path:   fmt.Sprintf("%s/%s/%s/items", baseHost, b.service.projectId, b.Name),
				Key:    b.service.key,
			}
			resp, err := req.do()
			if err != nil {
				panic(err)
			}
			respChannel <- newResponse(resp)
		}(body)
	}
	responses := make([]*Response, len(chunks))
	for i := 0; i < len(chunks); i++ {
		responses[i] = <-respChannel
	}
	return responses
}

func (b *base) fetch(last string) (*http.Response, error) {
	var body io.Reader
	path := fmt.Sprintf("%s/%s/%s/query", baseHost, b.service.projectId, b.Name)
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

// Get gets item(s) from the base associated with the given key(s).
// If no keys are provided, it returns the entire collection.
// Empty Get() might take a long time to return for large collections.
// If given keys are not found, it won't return any error.
// Check the Response to see if the item was found.
func (b *base) Get(keys ...string) []*Response {
	var container []*Response
	if len(keys) == 0 {
		var last string
		resp, err := b.fetch(last)
		if err != nil {
			panic(err)
		}
		cresp := newResponse(resp)
		container = append(container, cresp)
		if err != nil {
			panic(err)
		}
		lastValue, ok := cresp.Data["paging"].(map[string]interface{})["last"]
		if ok {
			last = lastValue.(string)
			for {
				resp, err := b.fetch(last)
				if err != nil {
					panic(err)
				}
				cresp := newResponse(resp)
				container = append(container, cresp)
				lastValue, ok := cresp.Data["paging"].(map[string]interface{})["last"]
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
			Path:   fmt.Sprintf("%s/%s/%s/items/%s", baseHost, b.service.projectId, b.Name, keys[0]),
			Key:    b.service.key,
		}
		resp, err := req.do()
		if err != nil {
			panic(err)
		}
		container = append(container, newResponse(resp))
	} else {
		responses := make(chan *Response, len(keys))
		for _, key := range keys {
			go func(key string) {
				req := httpRequest{
					Body:   nil,
					Method: "GET",
					Path:   fmt.Sprintf("%s/%s/%s/items/%s", baseHost, b.service.projectId, b.Name, key),
					Key:    b.service.key,
				}
				resp, err := req.do()
				if err != nil {
					responses <- nil
				}
				responses <- newResponse(resp)
			}(key)
		}
		for i := 0; i < len(keys); i++ {
			container = append(container, <-responses)
		}
	}
	return container
}

// Delete deletes item(s) from the collection.
// If no keys are provided, it returns an empty map[string]interface{}
// Even if given keys are not found, it won't return an error.
func (b *base) Delete(keys ...string) []*Response {
	respChannel := make(chan *Response, len(keys))
	for _, key := range keys {
		go func(key string) {
			req := httpRequest{
				Body:   nil,
				Method: "DELETE",
				Path:   fmt.Sprintf("%s/%s/%s/items/%s", baseHost, b.service.projectId, b.Name, key),
				Key:    b.service.key,
			}
			resp, _ := req.do()
			respChannel <- newResponse(resp)
		}(key)
	}
	resp := make([]*Response, len(keys))
	for i := 0; i < len(keys); i++ {
		resp[i] = <-respChannel
	}
	return resp
}

// Insert inserts a new item into the collection
// only if the item does not already exist.
// Array of Response is returned which contains the all data of the request.
func (b *base) Insert(items ...map[string]interface{}) []*Response {
	container := make(chan *Response, len(items))
	go func(item []map[string]interface{}) {
		for _, item := range items {
			reader, _ := interfaceReader(map[string]interface{}{"item": item})
			req := httpRequest{
				Body:   reader,
				Method: "POST",
				Path:   fmt.Sprintf("%s/%s/%s/items", baseHost, b.service.projectId, b.Name),
				Key:    b.service.key,
			}
			resp, err := req.do()
			if err != nil {
				panic(err)
			}
			container <- newResponse(resp)
		}
	}(items)
	resp := make([]*Response, len(items))
	for i := 0; i < len(items); i++ {
		resp[i] = <-container
	}
	return resp
}

// Update provides a way to update an item in the collection.
// If the item does not exist, it will not panic.
// Check the Response StatusCode to see if the item was found.
// It returns an updater object which can be used to update the item.
// updater has various update methods associated with it.
func (b *base) Update(key string) *updater {
	return &updater{
		key:      key,
		baseName: b.Name,
		service:  b.service,
		updates:  make(map[string]interface{}),
	}
}

// Fetch is used to do queries on the database.
// It returns a Response object which can be used to get the data.
// Response object contains a Data field which is a map[string]interface{}
// `last` is the last key for pagination and should be left empty for the first query.
// `limit` is the number of items to return per query, the maximum is 1000 and use 0 for the default.
func (b *base) Fetch(query *query, last string, limit int) *Response {
	var body []map[string]interface{}
	if len(query.ors) > 0 {
		body = query.ors
	} else if len(query.values) > 0 {
		body = append(body, query.values)
	} else {
		body = []map[string]interface{}{}
	}
	queryBody := map[string]interface{}{"query": body}
	if limit != 0 {
		queryBody["limit"] = limit
	}
	if last != "" {
		queryBody["last"] = last
	}
	reader, _ := interfaceReader(queryBody)
	req := httpRequest{
		Body:   reader,
		Method: "POST",
		Path:   fmt.Sprintf("%s/%s/%s/query", baseHost, b.service.projectId, b.Name),
		Key:    b.service.key,
	}
	resp, _ := req.do()
	return newResponse(resp)
}
