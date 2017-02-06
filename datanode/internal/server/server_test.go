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

	"github.com/apoydence/eachers/testhelpers"
	"github.com/apoydence/loggrebutterfly/datanode/internal/server"
	pb "github.com/apoydence/loggrebutterfly/api/v1"

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

	data             chan []byte
	errs             chan error
	mockWriteFetcher *mockWriteFetcher
	mockReadFetcher  *mockReadFetcher
}

func TestServer(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TS {
		mockWriteFetcher := newMockWriteFetcher()
		mockReadFetcher := newMockReadFetcher()

		data := make(chan []byte, 100)
		errs := make(chan error, 100)
		writer := func(d []byte) error {
			data <- d
			return <-errs
		}
		testhelpers.AlwaysReturn(mockWriteFetcher.WriterOutput.Writer, writer)

		close(mockWriteFetcher.WriterOutput.Err)
		close(mockReadFetcher.ReaderOutput.Err)

		addr, err := server.Start("127.0.0.1:0", mockWriteFetcher, mockReadFetcher)
		Expect(t, err == nil).To(BeTrue())

		return TS{
			T:                t,
			client:           fetchClient(addr),
			mockWriteFetcher: mockWriteFetcher,
			mockReadFetcher:  mockReadFetcher,
			data:             data,
			errs:             errs,
		}
	})

	o.Spec("it writes to the writer", func(t TS) {
		t.errs <- nil
		sender, err := t.client.Write(context.Background())
		Expect(t, err == nil).To(BeTrue())

		err = sender.Send(&pb.WriteInfo{
			Payload: []byte("some-data"),
		})
		Expect(t, err == nil).To(BeTrue())

		Expect(t, t.data).To(ViaPolling(
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
