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

func (u *updater) Do() (map[string]interface{}, error) {
	body, _ := interfaceReader(u.updates)
	req := httpRequest{
		Body:   body,
		Method: "PATCH",
		Path:   fmt.Sprintf("%s/%s/%s/items/%s", baseHost, u.service.projectId, u.baseName, u.key),
		Key:    u.service.key,
	}
	resp, err := req.do()
	if err != nil {
		return nil, err
	}
	return responseReader(resp)
}
