package filesystem

import (
	"context"
	"log"
	"time"

	"github.com/apoydence/petasos/router"
	"github.com/apoydence/talaria/pb"
	"google.golang.org/grpc"
)

type FileSystem struct {
	client pb.NodeClient
}

func New(addr string) *FileSystem {
	return &FileSystem{
		client: setupClient(addr),
	}
}

func (f *FileSystem) List() (file []string, err error) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	resp, err := f.client.ListClusters(ctx, new(pb.ListClustersInfo))
	if err != nil {
		return nil, err
	}

	return resp.Names, nil
}

func (f *FileSystem) Writer(name string) (writer router.Writer, err error) {
	sender, err := f.client.Write(context.Background())
	if err != nil {
		return nil, err
	}

	return nodeWriter{name: name, sender: sender}, nil
}

func setupClient(addr string) pb.NodeClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect to node: %v", err)
	}
	return pb.NewNodeClient(conn)
}

type nodeWriter struct {
	name   string
	sender pb.Node_WriteClient
}

func (w nodeWriter) Write(data []byte) (err error) {
	return w.sender.Send(&pb.WriteDataPacket{
		Name:    w.name,
		Message: data,
	})
}
