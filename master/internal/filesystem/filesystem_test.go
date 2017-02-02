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
	"github.com/apoydence/loggrebutterfly/master/internal/filesystem"
	"github.com/apoydence/onpar"
	pb "github.com/apoydence/talaria/api/v1"

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

type TF struct {
	*testing.T

	fs                  *filesystem.FileSystem
	mockSchedulerServer *mockSchedulerServer
}

func TestFileSystemCreate(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("when scheduler does not return an error", func() {
		o.BeforeEach(func(t TF) TF {
			testhelpers.AlwaysReturn(t.mockSchedulerServer.CreateOutput.Ret0, new(pb.CreateResponse))
			close(t.mockSchedulerServer.CreateOutput.Ret1)
			return t
		})

		o.Spec("it instructs the scheduler to create a buffer", func(t TF) {
			err := t.fs.Create("some-file")
			Expect(t, err == nil).To(BeTrue())

			Expect(t, t.mockSchedulerServer.CreateInput.Arg1).To(
				Chain(Receive(), Equal(&pb.CreateInfo{
					Name: "some-file",
				})),
			)
		})
	})
}

func TestFileSystemReadOnly(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("when scheduler does not return an error", func() {
		o.BeforeEach(func(t TF) TF {
			testhelpers.AlwaysReturn(t.mockSchedulerServer.ReadOnlyOutput.Ret0, new(pb.ReadOnlyResponse))
			close(t.mockSchedulerServer.ReadOnlyOutput.Ret1)
			return t
		})

		o.Spec("it instructs the scheduler to set the buffer to ReadOnly", func(t TF) {
			err := t.fs.ReadOnly("some-file")
			Expect(t, err == nil).To(BeTrue())

			Expect(t, t.mockSchedulerServer.ReadOnlyInput.Arg1).To(
				Chain(Receive(), Equal(&pb.ReadOnlyInfo{
					Name: "some-file",
				})),
			)
		})
	})
}

func TestFileSystemList(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("when scheduler does not return an error", func() {
		o.BeforeEach(func(t TF) TF {
			testhelpers.AlwaysReturn(t.mockSchedulerServer.ListClusterInfoOutput.Ret0, &pb.ListResponse{
				Info: []*pb.ClusterInfo{
					{Name: "a", Leader: "a"},
					{Name: "b", Leader: "b"},
					{Name: "c", Leader: "c"},
				},
			})
			close(t.mockSchedulerServer.ListClusterInfoOutput.Ret1)
			return t
		})

		o.Spec("it lists the buffers from the scheduler", func(t TF) {
			files, err := t.fs.List()
			Expect(t, err == nil).To(BeTrue())

			Expect(t, files).To(Contain("a", "b", "c"))
		})
	})
}

func TestFileSystemRoutes(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("when scheduler does not return an error", func() {
		o.BeforeEach(func(t TF) TF {
			testhelpers.AlwaysReturn(t.mockSchedulerServer.ListClusterInfoOutput.Ret0, &pb.ListResponse{
				Info: []*pb.ClusterInfo{
					{Name: "a", Leader: "a"},
					{Name: "b", Leader: "b"},
					{Name: "c", Leader: "c"},
				},
			})
			close(t.mockSchedulerServer.ListClusterInfoOutput.Ret1)
			return t
		})

		o.Spec("it lists the buffers from the scheduler", func(t TF) {
			routes, err := t.fs.Routes()
			Expect(t, err == nil).To(BeTrue())
			Expect(t, routes).To(Chain(HaveKey("a"), Equal("A")))
			Expect(t, routes).To(Chain(HaveKey("b"), Equal("B")))
			Expect(t, routes).To(Chain(HaveKey("c"), Equal("C")))
		})
	})
}

func setup(o *onpar.Onpar) {
	o.BeforeEach(func(t *testing.T) TF {
		addr, mockSchedulerServer := startMockSched()
		m := map[string]string{
			"a": "A",
			"b": "B",
			"c": "C",
		}

		return TF{
			T:                   t,
			mockSchedulerServer: mockSchedulerServer,
			fs:                  filesystem.New(addr, m),
		}
	})
}

func startMockSched() (string, *mockSchedulerServer) {
	mockSchedulerServer := newMockSchedulerServer()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSchedulerServer(s, mockSchedulerServer)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return lis.Addr().String(), mockSchedulerServer
}
