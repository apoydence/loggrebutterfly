package filesystem_test

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"

	"github.com/apoydence/eachers/testhelpers"
	"github.com/apoydence/loggrebutterfly/datanode/internal/filesystem"
	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
	"github.com/apoydence/talaria/pb"
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

	o.Group("when the node does not return an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			close(t.mockNodeServer.WriteOutput.Ret0)

			return t
		})

		o.Spec("it returns the list from the node", func(t TFS) {
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
