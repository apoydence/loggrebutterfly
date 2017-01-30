package server

import (
	"log"
	"net"

	"golang.org/x/net/context"

	"github.com/apoydence/loggrebutterfly/pb"

	"google.golang.org/grpc"
)

type Writer interface {
	Write(data []byte) (err error)
}

type ReadFetcher interface {
	Reader(name string) (reader func() ([]byte, error), err error)
}

type Server struct {
	writer Writer
	reader ReadFetcher
}

func Start(addr string, writer Writer, reader ReadFetcher) (actualAddr string, err error) {
	s := &Server{
		writer: writer,
		reader: reader,
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return "", err
	}
	g := grpc.NewServer()
	pb.RegisterDataNodeServer(g, s)

	go func() {
		if err := g.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	return lis.Addr().String(), nil
}

func (s *Server) Write(ctx context.Context, in *pb.WriteInfo) (*pb.WriteResponse, error) {
	if err := s.writer.Write(in.Payload); err != nil {
		return nil, err
	}

	return new(pb.WriteResponse), nil
}

func (s *Server) Read(info *pb.ReadInfo, server pb.DataNode_ReadServer) error {
	reader, err := s.reader.Reader(info.Name)
	if err != nil {
		return err
	}

	for {
		data, err := reader()
		if err != nil {
			return err
		}

		if err := server.Send(&pb.ReadData{Payload: data}); err != nil {
			return err
		}
	}
}
