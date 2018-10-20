package filesystem

import (
	"encoding/json"
	"log"

	"github.com/poy/petasos/router"
)

type Hasher interface {
	HashString(s string) (hash uint64)
}

type RouteFilter struct {
	hasher Hasher
}

func NewRouteFilter(h Hasher) *RouteFilter {
	return &RouteFilter{
		hasher: h,
	}
}

func (f *RouteFilter) Filter(route string, files map[string][]string) {
	hash := f.hasher.HashString(route)

	for file, _ := range files {
		if !f.inRange(file, hash) {
			delete(files, file)
		}
	}
}

func (f *RouteFilter) inRange(file string, hash uint64) bool {
	var rn router.RangeName
	if err := json.Unmarshal([]byte(file), &rn); err != nil {
		log.Printf("Error parsing file (%s) into RangeName: %s", file, err)
		return false
	}

	return rn.Low <= hash && rn.High >= hash
}
