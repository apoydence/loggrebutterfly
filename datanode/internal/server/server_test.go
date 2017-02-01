//go:generate hel

package server_test

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"google.golang.org/grpc"

	"github.com/apoydence/loggrebutterfly/datanode/internal/server"
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

type TS struct {
	*testing.T
	client pb.DataNodeClient

	mockWriter      *mockWriter
	mockReadFetcher *mockReadFetcher
}

func TestServer(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TS {
		mockWriter := newMockWriter()
		mockReadFetcher := newMockReadFetcher()
		close(mockWriter.WriteOutput.Err)

		close(mockReadFetcher.ReaderOutput.Err)

		addr, err := server.Start("127.0.0.1:0", mockWriter, mockReadFetcher)
		Expect(t, err == nil).To(BeTrue())

		return TS{
			T:               t,
			client:          fetchClient(addr),
			mockWriter:      mockWriter,
			mockReadFetcher: mockReadFetcher,
		}
	})

	o.Spec("it writes to the writer", func(t TS) {
		_, err := t.client.Write(context.Background(), &pb.WriteInfo{
			Payload: []byte("some-data"),
		})

		Expect(t, err == nil).To(BeTrue())
		Expect(t, t.mockWriter.WriteInput.Data).To(ViaPolling(
			Chain(Receive(), Equal([]byte("some-data"))),
		))
	})

	o.Spec("it reads from the reader", func(t TS) {
		t.mockReadFetcher.ReaderOutput.Reader <- buildDataF("A", "B", "C")

		resp, err := t.client.Read(context.Background(), &pb.ReadInfo{
			Name: "some-name",
		})
		Expect(t, err == nil).To(BeTrue())

		Expect(t, t.mockReadFetcher.ReaderInput.Name).To(ViaPolling(
			Chain(Receive(), Equal("some-name")),
		))

		data, err := resp.Recv()
		Expect(t, err == nil).To(BeTrue())
		Expect(t, data).To(Equal(&pb.ReadData{Payload: []byte("A")}))

		data, err = resp.Recv()
		Expect(t, err == nil).To(BeTrue())
		Expect(t, data).To(Equal(&pb.ReadData{Payload: []byte("B")}))

		data, err = resp.Recv()
		Expect(t, err == nil).To(BeTrue())
		Expect(t, data).To(Equal(&pb.ReadData{Payload: []byte("C")}))

		_, err = resp.Recv()
		Expect(t, err == nil).To(BeFalse())
	})
}

func fetchClient(addr string) pb.DataNodeClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return pb.NewDataNodeClient(conn)
}

func buildDataF(data ...string) func() ([]byte, error) {
	d := make(chan []byte, len(data)+1)
	e := make(chan error, len(data)+1)
	for _, x := range data {
		d <- []byte(x)
		e <- nil
	}
	d <- nil
	e <- io.EOF

	return func() ([]byte, error) {
		return <-d, <-e
	}
}
