package filesystem

import (
	"context"
	"log"
	"time"

	"github.com/apoydence/talaria/pb"
	"google.golang.org/grpc"
)

type FileSystem struct {
	schedClient pb.SchedulerClient
}

func New(addr string) *FileSystem {
	return &FileSystem{
		schedClient: setupSchedClient(addr),
	}
}

func (f *FileSystem) Create(file string) (err error) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	_, err = f.schedClient.Create(ctx, &pb.CreateInfo{Name: file})
	return err
}

func (f *FileSystem) ReadOnly(file string) (err error) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	_, err = f.schedClient.ReadOnly(ctx, &pb.ReadOnlyInfo{Name: file})
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
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	resp, err := f.schedClient.ListClusterInfo(ctx, new(pb.ListInfo))
	if err != nil {
		log.Printf("Failed to list cluster info: %s", err)
		return nil, err
	}

	for _, info := range resp.Info {
		routes[info.Name] = info.Leader
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
