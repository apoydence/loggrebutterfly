package intra

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/poy/loggrebutterfly/api/intra"
)

type Executor interface {
	Execute(fileName, algName string, ctx context.Context, meta []byte) (result map[string][]byte, err error)
}

type Server struct {
	exec Executor
}

func New(e Executor) *Server {
	return &Server{
		exec: e,
	}
}

func (s *Server) Execute(ctx context.Context, info *intra.ExecuteInfo) (resp *intra.ExecuteResponse, err error) {
	if info.File == "" {
		return nil, fmt.Errorf("File is required")
	}

	if info.Alg == "" {
		return nil, fmt.Errorf("Alg is required")
	}

	results, err := s.exec.Execute(info.File, info.Alg, ctx, info.Meta)
	if err != nil {
		return nil, err
	}

	return &intra.ExecuteResponse{
		Result: results,
	}, nil
}
