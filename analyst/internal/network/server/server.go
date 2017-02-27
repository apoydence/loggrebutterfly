package server

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strconv"

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
	if info.GetFilter().GetSourceId() == "" {
		return nil, fmt.Errorf("a source_id is required")
	}

	data, err := proto.Marshal(&v1.AggregateInfo{Query: info})
	if err != nil {
		return nil, err
	}

	result, err := s.calc.Calculate(info.GetFilter().GetSourceId(), "timerange", ctx, data)
	if err != nil {
		return nil, err
	}

	return &v1.QueryResponse{
		Envelopes: flattenResults(result),
	}, nil
}

func (s *Server) Aggregate(ctx context.Context, info *v1.AggregateInfo) (resp *v1.AggregateResponse, err error) {
	if info.GetQuery().GetFilter().GetSourceId() == "" {
		return nil, fmt.Errorf("a source_id is required")
	}

	if info.GetQuery().GetFilter().Envelopes == nil {
		return nil, fmt.Errorf("a envelope filter is required")
	}

	if info.BucketWidthNs == 0 {
		return nil, fmt.Errorf("a bucket_width_ns is required")
	}

	data, err := proto.Marshal(info)
	if err != nil {
		return nil, err
	}

	result, err := s.calc.Calculate(info.GetQuery().GetFilter().GetSourceId(), "aggregation", ctx, data)
	if err != nil {
		return nil, err
	}

	return &v1.AggregateResponse{
		Results: resultsToFloat(result),
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

func resultsToFloat(r map[string][]byte) map[int64]float64 {
	m := make(map[int64]float64)
	for k, v := range r {
		if len(v) != 8 {
			log.Printf("Invalid value (len=%d): %v", len(v), v)
			continue
		}

		bits := binary.LittleEndian.Uint64(v)
		float := math.Float64frombits(bits)
		i, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			log.Printf("Unable to parse key (%s) into int64: %s", k, err)
			continue
		}
		m[i] = float
	}

	return m
}
