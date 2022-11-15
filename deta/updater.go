package deta

import "fmt"

type updater struct {
	key      string
	baseName string
	service  *service
	updates  map[string]interface{}
}

func (u *updater) Set(attrs map[string]interface{}) {
	u.updates["set"] = attrs
}

func (u *updater) Delete(attrs ...string) {
	u.updates["delete"] = attrs
}

func (u *updater) Increment(attrs map[string]interface{}) {
	u.updates["increment"] = attrs
}

func (u *updater) Append(attrs map[string]interface{}) {
	u.updates["append"] = attrs
}

func (u *updater) Prepend(attrs map[string]interface{}) {
	u.updates["prepend"] = attrs
}

func (u *updater) Do() *response {
	body, _ := interfaceReader(u.updates)
	req := httpRequest{
		Body:   body,
		Method: "PATCH",
		Path:   fmt.Sprintf("%s/%s/%s/items/%s", baseHost, u.service.projectId, u.baseName, u.key),
		Key:    u.service.key,
	}
	resp, _ := req.do()
	return newResponse(resp)
}
