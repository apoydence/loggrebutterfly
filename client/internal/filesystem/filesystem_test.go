package filesystem_test

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/poy/eachers/testhelpers"
	pb "github.com/poy/loggrebutterfly/api/v1"
	"github.com/poy/loggrebutterfly/client/internal/filesystem"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
		grpclog.SetLogger(log.New(ioutil.Discard, "", 0))
	}

	os.Exit(m.Run())
}

type TFS struct {
	*testing.T
	dataNodeAddrs       []string
	mockDataNodeServers []*mockDataNodeServer
	dataNodeClients     []pb.DataNodeClient
	mockRouteCache      *mockRouteCache
	fs                  *filesystem.FileSystem
}

// TODO: These tests should not require an actual gRPC server

func TestFileSystemList(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("cache returns a client", func() {
		o.BeforeEach(func(t TFS) TFS {
			t.mockRouteCache.ListOutput.Ret0 <- []string{
				"some-name-a",
				"some-name-b",
			}
			return t
		})

		o.Spec("it lists the routes from the master", func(t TFS) {
			list, err := t.fs.List()
			Expect(t, err == nil).To(BeTrue())
			Expect(t, list).To(HaveLen(2))
			Expect(t, list).To(Contain(
				"some-name-a",
				"some-name-b",
			))
		})
	})
}

func TestFileSystemWriter(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("cache returns a client", func() {
		o.BeforeEach(func(t TFS) TFS {
			close(t.mockRouteCache.FetchRouteOutput.Addr)
			testhelpers.AlwaysReturn(t.mockRouteCache.FetchRouteOutput.Client, t.dataNodeClients[1])
			t.mockRouteCache.ListOutput.Ret0 <- []string{
				"some-name-a",
				"some-name-b",
			}
			return t
		})

		o.Group("data node does not return an error", func() {
			o.Spec("it writes to the correct data node", func(t TFS) {
				writer, err := t.fs.Writer("some-name-b")
				Expect(t, err == nil).To(BeTrue())
				err = writer.Write([]byte("some-data"))
				Expect(t, err == nil).To(BeTrue())

				var rx pb.DataNode_WriteServer
				Expect(t, t.mockDataNodeServers[1].WriteInput.Arg0).To(ViaPolling(
					Chain(Receive(), Fetch(&rx)),
				))
				data, err := rx.Recv()
				Expect(t, err == nil).To(BeTrue())
				Expect(t, data.Payload).To(Equal([]byte("some-data")))
			})
		})

		o.Group("when the data node returns an error", func() {
			o.BeforeEach(func(t TFS) TFS {
				testhelpers.AlwaysReturn(t.mockDataNodeServers[1].WriteOutput.Ret0, fmt.Errorf("some-error"))
				return t
			})

			o.Spec("it returns an error", func(t TFS) {
				writer, err := t.fs.Writer("some-name-b")
				Expect(t, err == nil).To(BeTrue())

				f := func() bool {
					return writer.Write([]byte("some-data")) != nil
				}
				Expect(t, f).To(ViaPolling(BeTrue()))
			})
		})
	})

	o.Group("cache does not return a client", func() {
		o.BeforeEach(func(t TFS) TFS {
			close(t.mockRouteCache.FetchRouteOutput.Addr)
			close(t.mockRouteCache.FetchRouteOutput.Client)
			t.mockRouteCache.ListOutput.Ret0 <- []string{
				"some-name-a",
				"some-name-b",
			}
			return t
		})

		o.Spec("it returns an error for an unknown file", func(t TFS) {
			_, err := t.fs.Writer("unknown")
			Expect(t, err == nil).To(BeFalse())
		})
	})
}

func TestFileSystemReader(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("when data node returns an EOF", func() {
		o.BeforeEach(func(t TFS) TFS {
			close(t.mockRouteCache.FetchRouteOutput.Addr)
			testhelpers.AlwaysReturn(t.mockRouteCache.FetchRouteOutput.Client, t.dataNodeClients[1])
			// t.mockRouteCache.ListOutput.Ret0 <- []string{
			// 	"some-name-a",
			// 	"some-name-b",
			// }
			testhelpers.AlwaysReturn(t.mockDataNodeServers[1].ReadOutput.Ret0, io.EOF)
			return t
		})

		o.Spec("it converts it to an io EOF", func(t TFS) {
			reader, err := t.fs.Reader("some-name-b", 99)
			Expect(t, err == nil).To(BeTrue())

			_, err = reader.Read()
			Expect(t, err).To(Equal(io.EOF))
		})
	})

	o.Group("when data node does not return an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			close(t.mockRouteCache.FetchRouteOutput.Addr)
			testhelpers.AlwaysReturn(t.mockRouteCache.FetchRouteOutput.Client, t.dataNodeClients[1])
			return t
		})

		o.Spec("it returns data from the data node", func(t TFS) {
			go func() {
				defer close(t.mockDataNodeServers[1].ReadOutput.Ret0)
				var rx pb.DataNode_ReadServer
				Expect(t, t.mockDataNodeServers[1].ReadInput.Arg1).To(ViaPolling(
					Chain(Receive(), Fetch(&rx)),
				))

				err := rx.Send(&pb.ReadData{
					Payload: []byte("some-data"),
					File:    "file-a",
					Index:   101,
				})
				Expect(t, err == nil).To(BeTrue())
			}()

			reader, err := t.fs.Reader("some-name-b", 99)
			Expect(t, err == nil).To(BeTrue())

			Expect(t, t.mockDataNodeServers[1].ReadInput.Arg0).To(ViaPolling(
				Chain(Receive(), Equal(&pb.ReadInfo{
					Name:  "some-name-b",
					Index: 99,
				})),
			))

			data, err := reader.Read()
			Expect(t, err == nil).To(BeTrue())
			Expect(t, data.Payload).To(Equal([]byte("some-data")))
			Expect(t, data.Filename).To(Equal("file-a"))
			Expect(t, data.Index).To(Equal(uint64(101)))
		})
	})
}

func setup(o *onpar.Onpar) {
	o.BeforeEach(func(t *testing.T) TFS {
		dataNodeAddrA, mockDataNodeServerA := startMockDataNode()
		dataNodeAddrB, mockDataNodeServerB := startMockDataNode()
		mockRouteCache := newMockRouteCache()

		return TFS{
			T:                   t,
			mockDataNodeServers: []*mockDataNodeServer{mockDataNodeServerA, mockDataNodeServerB},
			dataNodeAddrs:       []string{dataNodeAddrA, dataNodeAddrB},
			mockRouteCache:      mockRouteCache,
			dataNodeClients:     []pb.DataNodeClient{fetchDataNodeClient(dataNodeAddrA), fetchDataNodeClient(dataNodeAddrB)},
			fs:                  filesystem.New(mockRouteCache),
		}
	})
}

func startMockDataNode() (string, *mockDataNodeServer) {
	mockDataNodeServer := newMockDataNodeServer()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterDataNodeServer(s, mockDataNodeServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return lis.Addr().String(), mockDataNodeServer
}

func fetchDataNodeClient(addr string) pb.DataNodeClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return pb.NewDataNodeClient(conn)
}
