package intra

import (
	"encoding/json"
	"log"
	"net"

	"golang.org/x/net/context"

	"github.com/apoydence/loggrebutterfly/pb/intra"
	"github.com/apoydence/petasos/router"
	"google.golang.org/grpc"
)

type MetricsReader interface {
	Metrics(rn router.RangeName) (metric router.Metric)
}

type IntraServer struct {
	metricsReader MetricsReader
}

func Start(addr string, metricsReader MetricsReader) (actualAddr string, err error) {
	is := &IntraServer{metricsReader: metricsReader}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return "", err
	}

	s := grpc.NewServer()
	intra.RegisterDataNodeServer(s, is)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve (intra): %s", err)
		}
	}()

	return lis.Addr().String(), nil
}

func (s *IntraServer) ReadMetrics(ctx context.Context, in *intra.ReadMetricsInfo) (*intra.ReadMetricsResponse, error) {
	var rn router.RangeName
	if err := json.Unmarshal([]byte(in.File), &rn); err != nil {
		return nil, err
	}

	metrics := s.metricsReader.Metrics(rn)
	return &intra.ReadMetricsResponse{
		WriteCount: metrics.WriteCount,
		ErrCount:   metrics.ErrCount,
	}, nil
}
