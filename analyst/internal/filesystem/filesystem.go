package filesystem

import (
	"io"

	talaria "github.com/apoydence/talaria/api/v1"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type FileFilter interface {
	Filter(route string, files map[string][]string)
}

type FileSystem struct {
	filter      FileFilter
	schedClient talaria.SchedulerClient
	nodeClient  talaria.NodeClient

	toAnalyst   map[string]string
	fromAnalyst map[string]string
}

func New(f FileFilter, s talaria.SchedulerClient, n talaria.NodeClient, toAnalyst map[string]string) *FileSystem {
	return &FileSystem{
		filter:      f,
		schedClient: s,
		nodeClient:  n,
		toAnalyst:   toAnalyst,
	}
}

func (f *FileSystem) Files(route string, ctx context.Context, meta []byte) (files map[string][]string, err error) {
	files, err = f.fetchAllFiles(ctx)
	if err != nil {
		return nil, err
	}

	f.filter.Filter(route, files)
	f.swapToAnalyst(files)

	return files, nil
}

func (f *FileSystem) Reader(file string, ctx context.Context, meta []byte) (reader func() (data []byte, err error), err error) {
	rx, err := f.nodeClient.Read(ctx, &talaria.BufferInfo{Name: file})
	if err != nil {
		return nil, err
	}

	return func() ([]byte, error) {
		packet, err := rx.Recv()
		if grpc.ErrorDesc(err) == "EOF" {
			return nil, io.EOF
		}

		if err != nil {
			return nil, err
		}
		return packet.Message, nil
	}, nil
}

func (f *FileSystem) fetchAllFiles(ctx context.Context) (files map[string][]string, err error) {
	resp, err := f.schedClient.ListClusterInfo(ctx, new(talaria.ListInfo))
	if err != nil {
		return nil, err
	}

	files = make(map[string][]string)
	for _, info := range resp.Info {
		for _, nodeInfo := range info.Nodes {
			files[info.Name] = append(files[info.Name], nodeInfo.URI)
		}
	}

	return files, nil
}

func (f *FileSystem) swapToAnalyst(m map[string][]string) {
	for _, v := range m {
		for i := range v {
			v[i] = f.toAnalyst[v[i]]
		}
	}
}
