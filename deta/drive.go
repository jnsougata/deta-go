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
		var chunkedContent [][]byte
		for i := 0; i < chunks; i++ {
			start := i * maxChunkSize
			end := start + maxChunkSize
			if end > len(content) {
				end = len(content)
			}
			chunkedContent = append(chunkedContent, content[start:end])
		}
		lastChunk := chunkedContent[len(chunkedContent)-1]
		chunkedContent = chunkedContent[:len(chunkedContent)-1]
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
		restUploads := make(chan *Response, len(chunkedContent))
		for i, chunk := range chunkedContent {
			go func(i int, chunk []byte) {
				req := driveRequest{
					Body:   chunk,
					Method: "POST",
					Path: fmt.Sprintf(
						"%s/%s/%s/uploads/%s/parts?name=%s&part=%d",
						driveHost, d.service.projectId, d.Name, resp.UploadId, resp.Name, i+1),
					Key: d.service.key,
				}
				r, _ := req.Do()
				restUploads <- newResponse(r)
			}(i, chunk)
		}
		for i := 0; i < len(chunkedContent); i++ {
			<-restUploads
		}
		req := driveRequest{
			Body:   lastChunk,
			Method: "POST",
			Path: fmt.Sprintf(
				"%s/%s/%s/uploads/%s/parts?name=%s&part=%d",
				driveHost, d.service.projectId, d.Name, resp.UploadId, resp.Name, chunks),
			Key: d.service.key,
		}
		_, _ = req.Do()
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
