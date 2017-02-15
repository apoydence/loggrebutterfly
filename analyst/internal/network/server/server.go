package server

import (
	"fmt"
	"log"

	"golang.org/x/net/context"

	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
	"github.com/golang/protobuf/proto"
)

type Calculator interface {
	Calculate(route, algName string, ctx context.Context, meta []byte) (finalResult map[string][]byte, err error)
}

type Server struct {
	calc Calculator
}

func New(c Calculator) *Server {
	return &Server{
		calc: c,
	}
}

func (s *Server) Query(ctx context.Context, info *v1.QueryInfo) (resp *v1.QueryResponse, err error) {
	if info.SourceId == "" {
		return nil, fmt.Errorf("a SourceId is required")
	}

	data, err := proto.Marshal(info)
	if err != nil {
		return nil, err
	}

	result, err := s.calc.Calculate(info.SourceId, "timerange", ctx, data)
	if err != nil {
		return nil, err
	}

	return &v1.QueryResponse{
		Envelopes: flattenResults(result),
	}, nil
}

func flattenResults(m map[string][]byte) (results []*loggregator.Envelope) {
	for _, v := range m {
		var e loggregator.Envelope
		if err := proto.Unmarshal(v, &e); err != nil {
			log.Printf("Failed to unmarshal envelope: %s", err)
			continue
		}

		results = append(results, &e)
	}
	return results
}
