package algorithms

import (
	"fmt"

	v1 "github.com/apoydence/loggrebutterfly/api/v1"
	"github.com/apoydence/mapreduce"
	"github.com/golang/protobuf/proto"
)

type AlgBuilder func(req *v1.AggregateInfo) mapreduce.Algorithm

type Fetcher struct {
	builders map[string]AlgBuilder
}

func NewFetcher(m map[string]AlgBuilder) *Fetcher {
	return &Fetcher{
		builders: m,
	}
}

func (f *Fetcher) Alg(name string, meta []byte) (mapreduce.Algorithm, error) {
	builder, ok := f.builders[name]
	if !ok {
		return mapreduce.Algorithm{}, fmt.Errorf("unknown alg %s", name)
	}

	var req v1.AggregateInfo
	if err := proto.Unmarshal(meta, &req); err != nil {
		return mapreduce.Algorithm{}, err
	}
	return builder(&req), nil
}
