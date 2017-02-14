//go:generate hel

package filesystem_test

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"google.golang.org/grpc"

	"github.com/apoydence/loggrebutterfly/analyst/internal/filesystem"
	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
	talaria "github.com/apoydence/talaria/api/v1"
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
	mockFilter          *mockFileFilter
	mockSchedulerClient *mockSchedulerClient
	mockNodeClient      *mockNodeClient
	mockNodeReadClient  *mockNodeReadClient
	fs                  *filesystem.FileSystem
}

func TestFileSystemFiles(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("when the scheduler client does not return an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			t.mockSchedulerClient.ListClusterInfoOutput.Ret0 <- &talaria.ListResponse{
				Info: []*talaria.ClusterInfo{
					{
						Name: "a",
						Nodes: []*talaria.NodeInfo{
							{URI: "some-node-name-1"},
							{URI: "some-node-name-2"},
						},
					},
					{
						Name: "b",
						Nodes: []*talaria.NodeInfo{
							{URI: "some-node-name-1"},
							{URI: "some-node-name-3"},
						},
					},
					{
						Name: "c",
						Nodes: []*talaria.NodeInfo{
							{URI: "some-node-name-2"},
							{URI: "some-node-name-3"},
						},
					},
				},
			}

			close(t.mockSchedulerClient.ListClusterInfoOutput.Ret1)
			return t
		})

		o.Spec("it returns all the files within range of the route", func(t TFS) {
			files, err := t.fs.Files("some-route", context.Background(), nil)
			Expect(t, err == nil).To(BeTrue())
			Expect(t, files).To(HaveLen(3))

			Expect(t, files["a"]).To(Equal([]string{"translated-1", "translated-2"}))
			Expect(t, files["b"]).To(Equal([]string{"translated-1", "translated-3"}))
			Expect(t, files["c"]).To(Equal([]string{"translated-2", "translated-3"}))
		})

		o.Spec("it gives the file filter the expected route", func(t TFS) {
			t.fs.Files("some-route", context.Background(), nil)

			Expect(t, t.mockFilter.FilterInput.Route).To(
				Chain(Receive(), Equal("some-route")),
			)
		})

		o.Spec("it gives the range checker the expected files", func(t TFS) {
			t.fs.Files("some-route", context.Background(), nil)

			Expect(t, t.mockFilter.FilterInput.Files).To(
				Chain(Receive(), HaveLen(3)),
			)
		})
	})

	o.Group("when the scheduler returns an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			close(t.mockSchedulerClient.ListClusterInfoOutput.Ret0)
			t.mockSchedulerClient.ListClusterInfoOutput.Ret1 <- fmt.Errorf("some-error")
			return t
		})

		o.Spec("it returns an error", func(t TFS) {
			_, err := t.fs.Files("some-route", context.Background(), nil)
			Expect(t, err == nil).To(BeFalse())
		})
	})
}

func TestFileSystemReader(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	setup(o)

	o.Group("when the node client does not return an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			close(t.mockNodeClient.ReadOutput.Ret1)
			t.mockNodeClient.ReadOutput.Ret0 <- t.mockNodeReadClient
			return t
		})

		o.Group("node reader does not return an error", func() {
			o.BeforeEach(func(t TFS) TFS {
				close(t.mockNodeReadClient.RecvOutput.Ret1)
				t.mockNodeReadClient.RecvOutput.Ret0 <- &talaria.ReadDataPacket{
					Message: []byte("some-data"),
				}
				return t
			})

			o.Spec("it reads from the talaria client", func(t TFS) {
				reader, err := t.fs.Reader("some-file", context.Background(), nil)
				Expect(t, err == nil).To(BeTrue())

				data, err := reader()
				Expect(t, err == nil).To(BeTrue())
				Expect(t, data).To(Equal([]byte("some-data")))
				Expect(t, t.mockNodeClient.ReadInput.In).To(
					Chain(Receive(), Equal(&talaria.BufferInfo{Name: "some-file"})),
				)
			})
		})

		o.Group("node reader returns an error", func() {
			o.BeforeEach(func(t TFS) TFS {
				close(t.mockNodeReadClient.RecvOutput.Ret0)
				return t
			})

			o.Spec("it returns an error", func(t TFS) {
				t.mockNodeReadClient.RecvOutput.Ret1 <- fmt.Errorf("some-error")
				reader, err := t.fs.Reader("some-file", context.Background(), nil)
				Expect(t, err == nil).To(BeTrue())
				_, err = reader()
				Expect(t, err == nil).To(BeFalse())
			})

			o.Spec("it converts an EOF to a io EOF", func(t TFS) {
				t.mockNodeReadClient.RecvOutput.Ret1 <- grpc.Errorf(9, "EOF")
				reader, err := t.fs.Reader("some-file", context.Background(), nil)
				Expect(t, err == nil).To(BeTrue())
				_, err = reader()
				Expect(t, err).To(Equal(io.EOF))
			})
		})
	})

	o.Group("when the node client returns an error", func() {
		o.BeforeEach(func(t TFS) TFS {
			close(t.mockNodeClient.ReadOutput.Ret0)
			t.mockNodeClient.ReadOutput.Ret1 <- fmt.Errorf("some-error")
			return t
		})

		o.Spec("it returns an error", func(t TFS) {
			_, err := t.fs.Reader("some-file", context.Background(), nil)
			Expect(t, err == nil).To(BeFalse())
		})
	})
}

func setup(o *onpar.Onpar) {
	o.BeforeEach(func(t *testing.T) TFS {
		mockFilter := newMockFileFilter()
		mockSchedulerClient := newMockSchedulerClient()
		mockNodeClient := newMockNodeClient()
		mockNodeReadClient := newMockNodeReadClient()

		translate := map[string]string{
			"some-node-name-1": "translated-1",
			"some-node-name-2": "translated-2",
			"some-node-name-3": "translated-3",
		}

		return TFS{
			T:                   t,
			mockFilter:          mockFilter,
			mockSchedulerClient: mockSchedulerClient,
			mockNodeClient:      mockNodeClient,
			mockNodeReadClient:  mockNodeReadClient,
			fs:                  filesystem.New(mockFilter, mockSchedulerClient, mockNodeClient, translate),
		}
	})

}
