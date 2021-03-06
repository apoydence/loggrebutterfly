package server

import (
	"log"
	"net"

	pb "github.com/poy/loggrebutterfly/api/v1"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

type Lister interface {
	Routes() (routes map[string]string, err error)
}

type Server struct {
	lister       Lister
	analystAddrs []string
}

func Start(addr string, analystAddrs []string, lister Lister) (actualAddr string, err error) {
	s := &Server{
		lister:       lister,
		analystAddrs: analystAddrs,
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

func (s *Server) Analysts(ctx context.Context, in *pb.AnalystsInfo) (*pb.AnalystsResponse, error) {
	var info []*pb.AnalystInfo
	for _, addr := range s.analystAddrs {
		info = append(info, &pb.AnalystInfo{Addr: addr})
	}

	return &pb.AnalystsResponse{Analysts: info}, nil
}
