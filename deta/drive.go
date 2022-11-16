package deta

import (
	"encoding/json"
	"fmt"
)

const driveHost = "https://drive.deta.sh/v1"
const maxChunkSize = 1024 * 1024 * 10

type drive struct {
	Name    string
	service *service
}

// Put uploads the file with the given name.
// If the file already exists, it is overwritten.
func (d *drive) Put(name string, content []byte) *Response {
	if len(content) <= maxChunkSize {
		req := driveRequest{
			Body:   content,
			Method: "POST",
			Path:   fmt.Sprintf("%s/%s/%s/files?name=%s", driveHost, d.service.projectId, d.Name, name),
			Key:    d.service.key,
		}
		resp, err := req.Do()
		if err != nil {
			panic(err)
		}
		return newResponse(resp)
	} else {
		chunks := len(content) / maxChunkSize
		if len(content)%maxChunkSize != 0 {
			chunks++
		}
		var parts [][]byte
		for i := 0; i < chunks; i++ {
			start := i * maxChunkSize
			end := start + maxChunkSize
			if end > len(content) {
				end = len(content)
			}
			parts = append(parts, content[start:end])
		}
		initiateReq := driveRequest{
			Method: "POST",
			Path:   fmt.Sprintf("%s/%s/%s/uploads?name=%s", driveHost, d.service.projectId, d.Name, name),
			Key:    d.service.key,
		}
		initiateResp, err := initiateReq.Do()
		if err != nil {
			panic(err)
		}
		var resp struct {
			Name      string `json:"name"`
			UploadId  string `json:"upload_id"`
			ProjectId string `json:"project_id"`
			DriveName string `json:"drive_name"`
		}
		err = json.NewDecoder(initiateResp.Body).Decode(&resp)
		if err != nil {
			panic(err)
		}
		uploads := make(chan *Response, len(parts))
		for i, part := range parts {
			go func(i int, part []byte) {
				req := driveRequest{
					Body:   part,
					Method: "POST",
					Path: fmt.Sprintf(
						"%s/%s/%s/uploads/%s/parts?name=%s&part=%d",
						driveHost, d.service.projectId, d.Name, resp.UploadId, resp.Name, i+1),
					Key: d.service.key,
				}
				r, _ := req.Do()
				uploads <- newResponse(r)
			}(i, part)
		}
		for i := 0; i < len(parts); i++ {
			<-uploads
		}
		completeReq := driveRequest{
			Method: "PATCH",
			Path: fmt.Sprintf(
				"%s/%s/%s/uploads/%s?name=%s",
				driveHost, d.service.projectId, d.Name, resp.UploadId, resp.Name),
			Key: d.service.key,
		}
		finalResp, err := completeReq.Do()
		if err != nil {
			panic(err)
		}
		return newResponse(finalResp)
	}
}

// Get returns the file as ReadCloser with the given name.
func (d *drive) Get(name string) *StreamingResponse {
	req := driveRequest{
		Method: "GET",
		Path:   fmt.Sprintf("%s/%s/%s/files/download?name=%s", driveHost, d.service.projectId, d.Name, name),
		Key:    d.service.key,
	}
	resp, err := req.Do()
	if err != nil {
		panic(err)
	}
	return &StreamingResponse{resp.StatusCode, resp.Body}
}

// Delete deletes the files with the given names.
func (d *drive) Delete(names ...string) *Response {
	body, _ := interfaceReader(map[string][]string{"names": names})
	req := httpRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("%s/%s/%s/files", driveHost, d.service.projectId, d.Name),
		Key:    d.service.key,
		Body:   body,
	}
	resp, err := req.Do()
	if err != nil {
		panic(err)
	}
	return newResponse(resp)
}

// Files returns all the files in the drive with the given prefix.
// If prefix is empty, all files are returned.
// limit <- the number of files to return, defaults to 1000.
// last <- last filename of the previous request to get the next set of files.
// Use limit 0 and last "" to obtain the default behaviour of the drive.
func (d *drive) Files(prefix string, limit int, last string) *Response {
	if limit == 0 || limit > 1000 || limit < 0 {
		limit = 1000
	}
	path := fmt.Sprintf("%s/%s/%s/files?limit=%d", driveHost, d.service.projectId, d.Name, limit)
	if prefix != "" {
		path += fmt.Sprintf("&prefix=%s", prefix)
	}
	if last != "" {
		path += fmt.Sprintf("&last=%s", last)
	}
	req := httpRequest{
		Method: "GET",
		Path:   path,
		Key:    d.service.key,
	}
	resp, err := req.Do()
	if err != nil {
		panic(err)
	}
	return newResponse(resp)
}
