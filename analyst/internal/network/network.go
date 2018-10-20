package network

import (
	"io"
	"log"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/poy/loggrebutterfly/api/intra"
)

type Network struct {
}

func New() *Network {
	return &Network{}
}

func (n *Network) Execute(file, algName, nodeID string, ctx context.Context, meta []byte) (result map[string][]byte, err error) {
	client, closer := setupClient(nodeID)
	defer closer.Close()

	resp, err := client.Execute(ctx, &intra.ExecuteInfo{
		File: file,
		Alg:  algName,
		Meta: meta,
	})
	if err != nil {
		return nil, err
	}

	return resp.Result, nil
}

func setupClient(addr string) (intra.AnalystClient, io.Closer) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect to analyst: %s", err)
	}
	return intra.NewAnalystClient(conn), conn
}
