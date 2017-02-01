package server

import (
	"log"
	"net"

	pb "github.com/apoydence/loggrebutterfly/pb/v1"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

type Lister interface {
	Routes() (routes map[string]string, err error)
}

type Server struct {
	lister Lister
}

func Start(addr string, lister Lister) (actualAddr string, err error) {
	s := &Server{
		lister: lister,
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return "", nil
	}
	g := grpc.NewServer()
	pb.RegisterMasterServer(g, s)
	go func() {
		if err := g.Serve(lis); err != nil {
			log.Fatalf("unable to serve: %v", err)
		}
	}()
	return lis.Addr().String(), nil
}

func (s *Server) Routes(ctx context.Context, in *pb.RoutesInfo) (*pb.RoutesResponse, error) {
	routes, err := s.lister.Routes()
	if err != nil {
		return nil, err
	}

	var resp pb.RoutesResponse
	for route, leader := range routes {
		resp.Routes = append(resp.Routes, &pb.RouteInfo{
			Name:   route,
			Leader: leader,
		})
	}

	return &resp, nil
}
