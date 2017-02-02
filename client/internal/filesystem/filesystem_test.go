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

	"github.com/apoydence/eachers/testhelpers"
	"github.com/apoydence/loggrebutterfly/client/internal/filesystem"
	pb "github.com/apoydence/loggrebutterfly/pb/v1"
	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	os.Exit(m.Run())
}

type TFS struct {
	*testing.T
	dataNodeAddrs       []string
	mockMasterServer    *mockMasterServer
	mockDataNodeServers []*mockDataNodeServer
	fs                  *filesystem.FileSystem
}

func TestFileSystemList(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("master does not return an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			writeRoutes(t)
			return t
		})

		o.Spec("it lists the routes from the master", func(t TFS) {
			list, err := t.fs.List()
			Expect(t, err == nil).To(BeTrue())
			Expect(t, list).To(HaveLen(2))
			Expect(t, list[0]).To(Or(
				Equal("some-name-a"),
				Equal("some-name-b"),
			))
			Expect(t, list[1]).To(Or(
				Equal("some-name-a"),
				Equal("some-name-b"),
			))
			Expect(t, list[0]).To(Not(Equal(list[1])))
		})

		o.Spec("it does not query the master each time", func(t TFS) {
			t.fs.List()
			t.fs.List()

			Expect(t, t.mockMasterServer.RoutesCalled).To(Always(HaveLen(1)))
		})
	})
}

func TestFileSystemWriter(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("master does not return an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			writeRoutes(t)
			return t
		})

		o.Group("data node does not return an error", func() {
			o.BeforeEach(func(t TFS) TFS {
				testhelpers.AlwaysReturn(t.mockDataNodeServers[1].WriteOutput.Ret0, new(pb.WriteResponse))
				close(t.mockDataNodeServers[1].WriteOutput.Ret1)
				return t
			})

			o.Spec("it writes to the correct data node", func(t TFS) {
				writer, err := t.fs.Writer("some-name-b")
				Expect(t, err == nil).To(BeTrue())
				err = writer.Write([]byte("some-data"))
				Expect(t, err == nil).To(BeTrue())

				var info *pb.WriteInfo
				Expect(t, t.mockDataNodeServers[1].WriteInput.Arg1).To(ViaPolling(
					Chain(Receive(), Fetch(&info)),
				))
				Expect(t, info.Payload).To(Equal([]byte("some-data")))
			})
		})

		o.Group("when the data node returns an error", func() {
			o.BeforeEach(func(t TFS) TFS {
				close(t.mockDataNodeServers[1].WriteOutput.Ret0)
				testhelpers.AlwaysReturn(t.mockDataNodeServers[1].WriteOutput.Ret1, fmt.Errorf("some-error"))
				return t
			})

			o.Spec("it returns an error", func(t TFS) {
				writer, err := t.fs.Writer("some-name-b")
				Expect(t, err == nil).To(BeTrue())

				err = writer.Write([]byte("some-data"))
				Expect(t, err == nil).To(BeFalse())
			})

			o.Spec("it does not write to the data node once it is dead", func(t TFS) {
				writer, err := t.fs.Writer("some-name-b")
				Expect(t, err == nil).To(BeTrue())

				err = writer.Write([]byte("some-data"))
				Expect(t, err == nil).To(BeFalse())

				err = writer.Write([]byte("some-data"))
				Expect(t, err == nil).To(BeFalse())

				Expect(t, t.mockDataNodeServers[1].WriteCalled).To(Always(HaveLen(1)))
			})

			o.Spec("it refreshes the routes after an error", func(t TFS) {
				writer, err := t.fs.Writer("some-name-b")
				Expect(t, err == nil).To(BeTrue())

				err = writer.Write([]byte("some-data"))
				Expect(t, err == nil).To(BeFalse())

				writer, err = t.fs.Writer("some-name-b")
				Expect(t, err == nil).To(BeTrue())

				Expect(t, t.mockMasterServer.RoutesCalled).To(ViaPolling(HaveLen(2)))
			})
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

	o.Group("when data node returns an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			writeRoutes(t)
			testhelpers.AlwaysReturn(t.mockDataNodeServers[1].ReadOutput.Ret0, io.EOF)
			return t
		})

		o.Spec("it converts it to an io.EOF", func(t TFS) {
			reader, err := t.fs.Reader("some-name-b")
			Expect(t, err == nil).To(BeTrue())

			_, err = reader.Read()
			Expect(t, err).To(Equal(io.EOF))
		})
	})

	o.Group("when data node does not return an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			writeRoutes(t)
			return t
		})

		o.Spec("it returns data from the data node", func(t TFS) {
			go func() {
				defer close(t.mockDataNodeServers[1].ReadOutput.Ret0)
				var rx pb.DataNode_ReadServer
				Expect(t, t.mockDataNodeServers[1].ReadInput.Arg1).To(ViaPolling(
					Chain(Receive(), Fetch(&rx)),
				))

				err := rx.Send(&pb.ReadData{Payload: []byte("some-data")})
				Expect(t, err == nil).To(BeTrue())
			}()

			reader, err := t.fs.Reader("some-name-b")
			Expect(t, err == nil).To(BeTrue())

			Expect(t, t.mockDataNodeServers[1].ReadInput.Arg0).To(ViaPolling(
				Chain(Receive(), Equal(&pb.ReadInfo{Name: "some-name-b"})),
			))

			data, err := reader.Read()
			Expect(t, err == nil).To(BeTrue())
			Expect(t, data).To(Equal([]byte("some-data")))
		})
	})
}

func writeRoutes(t TFS) {
	testhelpers.AlwaysReturn(t.mockMasterServer.RoutesOutput.Ret0, &pb.RoutesResponse{
		Routes: []*pb.RouteInfo{
			{
				Name:   "some-name-a",
				Leader: t.dataNodeAddrs[0],
			},
			{
				Name:   "some-name-b",
				Leader: t.dataNodeAddrs[1],
			},
		},
	})
	close(t.mockMasterServer.RoutesOutput.Ret1)
}

func setup(o *onpar.Onpar) {
	o.BeforeEach(func(t *testing.T) TFS {
		masterAddr, mockMasterServer := startMockMaster()
		dataNodeAddrA, mockDataNodeServerA := startMockDataNode()
		dataNodeAddrB, mockDataNodeServerB := startMockDataNode()

		return TFS{
			T:                   t,
			mockMasterServer:    mockMasterServer,
			mockDataNodeServers: []*mockDataNodeServer{mockDataNodeServerA, mockDataNodeServerB},
			dataNodeAddrs:       []string{dataNodeAddrA, dataNodeAddrB},
			fs:                  filesystem.New(masterAddr),
		}
	})
}

func startMockMaster() (string, *mockMasterServer) {
	mockMasterServer := newMockMasterServer()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterMasterServer(s, mockMasterServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return lis.Addr().String(), mockMasterServer
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
