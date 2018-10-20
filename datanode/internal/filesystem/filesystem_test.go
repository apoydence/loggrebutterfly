package filesystem_test

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"

	"github.com/poy/eachers/testhelpers"
	"github.com/poy/loggrebutterfly/datanode/internal/filesystem"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
	pb "github.com/poy/talaria/api/v1"
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
	fs             *filesystem.FileSystem
	mockNodeServer *mockNodeServer
}

func TestFileSystemList(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("when the node does not return an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			testhelpers.AlwaysReturn(t.mockNodeServer.ListClustersOutput.Ret0, &pb.ListClustersResponse{
				Names: []string{"a", "b", "c"},
			})
			close(t.mockNodeServer.ListClustersOutput.Ret1)
			return t
		})

		o.Spec("it returns the list from the node", func(t TFS) {
			list, err := t.fs.List()
			Expect(t, err == nil).To(BeTrue())

			Expect(t, list).To(Contain("a", "b", "c"))
		})
	})
}

func TestFileSystemWriter(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Spec("it returns the list from the node", func(t TFS) {
		defer close(t.mockNodeServer.WriteOutput.Ret0)
		go func() {
			writer, err := t.fs.Writer("some-name")
			Expect(t, err == nil).To(BeTrue())

			err = writer.Write([]byte("some-data"))
			Expect(t, err == nil).To(BeTrue())
		}()

		var writer pb.Node_WriteServer
		Expect(t, t.mockNodeServer.WriteInput.Arg0).To(ViaPolling(
			Chain(Receive(), Fetch(&writer)),
		))

		packet, err := writer.Recv()
		Expect(t, err == nil).To(BeTrue())
		Expect(t, packet.Message).To(Equal([]byte("some-data")))
	})
}

func TestFileSystemReader(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Spec("it returns data from the node", func(t TFS) {
		defer close(t.mockNodeServer.ReadOutput.Ret0)
		go func() {
			var rx pb.Node_ReadServer
			Expect(t, t.mockNodeServer.ReadInput.Arg1).To(ViaPolling(
				Chain(Receive(), Fetch(&rx)),
			))

			err := rx.Send(&pb.ReadDataPacket{
				Message: []byte("some-data"),
				Index:   99,
			})
			Expect(t, err == nil).To(BeTrue())
		}()

		reader, err := t.fs.Reader("some-name", 99)
		Expect(t, err == nil).To(BeTrue())

		data, err := reader()
		Expect(t, err == nil).To(BeTrue())
		Expect(t, data.Payload).To(Equal([]byte("some-data")))
		Expect(t, data.File).To(Equal("some-name"))
		Expect(t, data.Index).To(Equal(uint64(99)))

		Expect(t, t.mockNodeServer.ReadInput.Arg0).To(ViaPolling(
			Chain(Receive(), Equal(&pb.BufferInfo{
				Name:       "some-name",
				StartIndex: 99,
			})),
		))
	})
}

func setup(o *onpar.Onpar) {
	o.BeforeEach(func(t *testing.T) TFS {
		addr, mockNodeServer := startMockNode()
		return TFS{
			T:              t,
			fs:             filesystem.New(addr),
			mockNodeServer: mockNodeServer,
		}
	})
}

func startMockNode() (string, *mockNodeServer) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	mockNodeServer := newMockNodeServer()
	s := grpc.NewServer()
	pb.RegisterNodeServer(s, mockNodeServer)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return lis.Addr().String(), mockNodeServer
}
