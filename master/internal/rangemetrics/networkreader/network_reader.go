package networkreader

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/apoydence/loggrebutterfly/pb/intra"
	"github.com/apoydence/petasos/router"
)

type NetworkReader struct {
	mu      sync.Mutex
	clients map[string]intra.DataNodeClient
}

func New() *NetworkReader {
	return &NetworkReader{
		clients: make(map[string]intra.DataNodeClient),
	}
}

func (r *NetworkReader) ReadMetrics(addr, file string) (metric router.Metric, err error) {
	client, err := r.fetchClient(addr)
	if err != nil {
		return router.Metric{}, err
	}

	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	resp, err := client.ReadMetrics(ctx, &intra.ReadMetricsInfo{File: file})
	if err != nil {
		return router.Metric{}, err
	}

	return router.Metric{
		WriteCount: resp.WriteCount,
		ErrCount:   resp.ErrCount,
	}, nil
}

func (r *NetworkReader) fetchClient(addr string) (intra.DataNodeClient, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, ok := r.clients[addr]
	if ok {
		return client, nil
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client = intra.NewDataNodeClient(conn)
	r.clients[addr] = client

	return client, nil
}
