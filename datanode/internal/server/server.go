package server

import (
	"log"
	"net"

	pb "github.com/apoydence/loggrebutterfly/api/v1"

	"google.golang.org/grpc"
)

type WriteFetcher interface {
	Writer() (writer func(data []byte) (err error), err error)
}

type ReadFetcher interface {
	Reader(name string) (reader func() ([]byte, error), err error)
}

type Server struct {
	writer WriteFetcher
	reader ReadFetcher
}

func Start(addr string, writer WriteFetcher, reader ReadFetcher) (actualAddr string, err error) {
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

func (s *Server) Write(sender pb.DataNode_WriteServer) error {
	w, err := s.writer.Writer()
	if err != nil {
		return err
	}

	for {
		info, err := sender.Recv()
		if err != nil {
			return err
		}

		if err := w(info.Payload); err != nil {
			return err
		}
	}
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
