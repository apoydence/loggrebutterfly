package filesystem

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/apoydence/loggrebutterfly/pb"
	"github.com/apoydence/petasos/router"
	"google.golang.org/grpc"
)

type FileSystem struct {
	masterClient pb.MasterClient

	routes map[string]clientInfo
}

type clientInfo struct {
	client pb.DataNodeClient
	closer io.Closer
}

func New(masterAddr string) *FileSystem {
	return &FileSystem{
		masterClient: setupMasterClient(masterAddr),
	}
}

func (f *FileSystem) List() (files []string, err error) {
	if err := f.setupRoutes(); err != nil {
		return nil, err
	}

	for name, _ := range f.routes {
		files = append(files, name)
	}

	return files, nil
}

func (f *FileSystem) Writer(name string) (writer router.Writer, err error) {
	client, ok := f.fetchRoute(name)
	if !ok {
		return nil, fmt.Errorf("unknown file: %s", name)
	}

	wrapper := &senderWrapper{
		client: client,
		reset: func() {
			r := f.routes
			f.routes = nil

			for _, client := range r {
				client.closer.Close()
			}
		},
	}

	return wrapper, nil
}

func (f *FileSystem) setupRoutes() error {
	if f.routes != nil {
		return nil
	}

	files, err := f.list()
	if err != nil {
		return err
	}

	f.routes = make(map[string]clientInfo)

	for _, file := range files {
		f.routes[file.Name] = setupDataClient(file.Leader)
	}

	return nil
}

func (f *FileSystem) fetchRoute(name string) (client pb.DataNodeClient, ok bool) {
	if err := f.setupRoutes(); err != nil {
		return nil, false
	}

	info, ok := f.routes[name]
	return info.client, ok
}

func (f *FileSystem) list() (files []*pb.RouteInfo, err error) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	resp, err := f.masterClient.Routes(ctx, new(pb.RoutesInfo))
	if err != nil {
		return nil, err
	}

	return resp.Routes, nil
}

func setupMasterClient(addr string) pb.MasterClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to master: %s", err)
	}
	return pb.NewMasterClient(conn)
}

func setupDataClient(addr string) clientInfo {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to master: %s", err)
	}
	return clientInfo{
		client: pb.NewDataNodeClient(conn),
		closer: conn,
	}
}

type senderWrapper struct {
	client pb.DataNodeClient
	err    error
	reset  func()
}

func (w *senderWrapper) Write(data []byte) error {
	if w.err != nil {
		return w.err
	}

	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
	_, w.err = w.client.Write(ctx, &pb.WriteInfo{Payload: data})

	if w.err != nil {
		w.reset()
		return w.err
	}

	return nil
}
