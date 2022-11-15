package deta

import (
	"os"
	"strings"
)


type deta struct {
	service *service
}

// Base returns the pointer to a new base instance
func (d *deta) Base(name string) *base {
	return &base{Name: name, service: d.service}
}

// Drive returns the pointer to a new drive instance
func (d *deta) Drive(name string) *drive {
	return &drive{Name: name, service: d.service}
}

// New returns the pointer to a new deta instance.
// args is used to pass the project key as a string.
// Only the first argument is used and the rest are ignored. 
// Variadic arguments are used to make it easier to optionally pass the project key.
// If not passed, it will try to read from the environment variable DETA_PROJECT_KEY
func New(args ...string) *deta {
	var key string
	if len(args) > 0 {
		key = args[0]
	} else {
		key = os.Getenv("DETA_PROJECT_KEY")
	}

	fragments := strings.Split(key, "_")
	if len(fragments) != 2 {
		panic("invalid project key is given, visit https://web.deta.sh")
	}
	service := service{key: key, projectId: fragments[0]}
	return &deta{service: &service}
}