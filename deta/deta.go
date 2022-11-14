package deta

import (
	"os"
	"strings"
)


type Deta struct {
	service *service
}

func (d *Deta) Base(name string) *base {
	return &base{Name: name, service: d.service}
}

func (d *Deta) Drive(name string) *drive {
	return &drive{Name: name, service: d.service}
}

func New(args ...string) *Deta {
	var key string
	if len(args) > 0 {
		key = args[0]
	} else {
		key = os.Getenv("DETA_PROJECT_KEY")
	}

	fragments := strings.Split(key, "_")
	if len(fragments) != 2 {
		panic("invalid project key, expected format >> id_key")
	}
	service := service{key: key, projectId: fragments[0]}
	return &Deta{service: &service}
}