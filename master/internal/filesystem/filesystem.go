package filesystem

import (
	"context"
	"log"
	"time"

	pb "github.com/apoydence/talaria/api/v1"
	"google.golang.org/grpc"
)

type FileSystem struct {
	schedClient       pb.SchedulerClient
	nodeAddrConverter map[string]string
}

func New(addr string, nodeAddrConverter map[string]string) *FileSystem {
	return &FileSystem{
		schedClient:       setupSchedClient(addr),
		nodeAddrConverter: nodeAddrConverter,
	}
}

func (f *FileSystem) Create(file string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = f.schedClient.Create(ctx, &pb.CreateInfo{Name: file})
	return err
}

func (f *FileSystem) List() (files []string, err error) {
	m, err := f.Routes()
	if err != nil {
		return nil, err
	}

	for name, _ := range m {
		files = append(files, name)
	}

	return files, nil
}

func (f *FileSystem) Routes() (routes map[string]string, err error) {
	routes = make(map[string]string)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := f.schedClient.ListClusterInfo(ctx, new(pb.ListInfo))
	if err != nil {
		log.Printf("Failed to list cluster info: %s", err)
		return nil, err
	}

	for _, info := range resp.Info {
		addr, ok := f.nodeAddrConverter[info.Leader]
		if !ok {
			log.Printf("Unknown node address (info=%+v) (name=%s): '%s'\n", info, info.Name, info.Leader)
		}
		routes[info.Name] = addr
	}

	return routes, nil
}

func setupSchedClient(addr string) (client pb.SchedulerClient) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to scheduler: %v", err)
	}
	return pb.NewSchedulerClient(conn)
}
