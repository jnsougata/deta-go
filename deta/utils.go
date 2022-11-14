package deta

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func interfaceReader(data interface{}) (io.Reader, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body), nil
}

func autoSlice(data []map[string]interface{}, length int) [][]map[string]interface{} {
	var slices [][]map[string]interface{}
	for i := 0; i < len(data); i += length {
		end := i + length
		if end > len(data) {
			end = len(data)
		}
		slices = append(slices, data[i:end])
	}
	return slices
}

func responseReader(resp *http.Response) (map[string]interface{}, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	return data, nil
}
