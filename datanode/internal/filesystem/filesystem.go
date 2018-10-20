package filesystem

import (
	"context"
	"log"
	"time"

	v1 "github.com/poy/loggrebutterfly/api/v1"
	"github.com/poy/petasos/router"
	pb "github.com/poy/talaria/api/v1"
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
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := f.client.ListClusters(ctx, new(pb.ListClustersInfo))
	if err != nil {
		return nil, err
	}

	return resp.Names, nil
}

func (f *FileSystem) Writer(name string) (writer router.Writer, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	sender, err := f.client.Write(ctx)
	if err != nil {
		return nil, err
	}

	return nodeWriter{name: name, sender: sender}, nil
}

func (f *FileSystem) Reader(name string, startIndex uint64) (reader func() (*v1.ReadData, error), err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	rx, err := f.client.Read(ctx, &pb.BufferInfo{Name: name, StartIndex: startIndex})
	if err != nil {
		return nil, err
	}

	return func() (*v1.ReadData, error) {
		packet, err := rx.Recv()
		if err != nil {
			return nil, err
		}

		return &v1.ReadData{
			Payload: packet.Message,
			File:    name,
			Index:   packet.Index,
		}, nil
	}, nil
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

func (w nodeWriter) Close() {
	w.sender.CloseAndRecv()
}
