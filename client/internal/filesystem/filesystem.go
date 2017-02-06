package filesystem

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"

	pb "github.com/apoydence/loggrebutterfly/api/v1"
	"github.com/apoydence/petasos/reader"
	"github.com/apoydence/petasos/router"
)

type RouteCache interface {
	List() []string
	FetchRoute(name string) (client pb.DataNodeClient, addr string)
	Reset()
}

type FileSystem struct {
	cache RouteCache
}

func New(cache RouteCache) *FileSystem {
	return &FileSystem{
		cache: cache,
	}
}

func (f *FileSystem) List() (files []string, err error) {
	list := f.cache.List()
	if len(list) == 0 {
		return nil, fmt.Errorf("unable to fetch route list")
	}
	return list, nil
}

func (f *FileSystem) Writer(name string) (writer router.Writer, err error) {
	client, addr := f.cache.FetchRoute(name)
	if client == nil {
		return nil, fmt.Errorf("unknown file: %s", name)
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	sender, err := client.Write(ctx)
	if err != nil {
		f.cache.Reset()
		return nil, err
	}

	wrapper := &senderWrapper{
		sender: sender,
		addr:   addr,
		reset:  f.cache.Reset,
	}

	return wrapper, nil
}

func (f *FileSystem) Reader(name string) (reader reader.Reader, err error) {
	client, addr := f.cache.FetchRoute(name)
	if client == nil {
		return nil, fmt.Errorf("unknown file: %s", name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	rx, err := client.Read(ctx, &pb.ReadInfo{Name: name})
	if err != nil {
		return nil, err
	}

	return &receiverWrapper{rx: rx, cancel: cancel, addr: addr}, nil
}

type senderWrapper struct {
	addr   string
	sender pb.DataNode_WriteClient
	err    error
	reset  func()
}

func (w *senderWrapper) Write(data []byte) error {
	if w.err != nil {
		return w.err
	}

	err := w.sender.Send(&pb.WriteInfo{Payload: data})
	if err != nil {
		w.err = err
		w.reset()
		return fmt.Errorf("[WRITE TO %s]: %s", w.addr, w.err)
	}

	return nil
}

type receiverWrapper struct {
	addr   string
	rx     pb.DataNode_ReadClient
	cancel func()
}

func (w *receiverWrapper) Read() ([]byte, error) {
	data, err := w.rx.Recv()
	if err != nil && grpc.ErrorDesc(err) == "EOF" {
		return nil, io.EOF
	}

	if err != nil {
		return nil, fmt.Errorf("[READ FROM %s]: %s", w.addr, err)
	}

	return data.Payload, nil
}

func (w *receiverWrapper) Close() {
	w.cancel()
}
