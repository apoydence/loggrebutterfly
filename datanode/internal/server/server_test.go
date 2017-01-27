//go:generate hel

package server_test

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"google.golang.org/grpc"

	"github.com/apoydence/loggrebutterfly/datanode/internal/server"
	"github.com/apoydence/loggrebutterfly/pb"

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
	client     pb.DataNodeClient
	mockWriter *mockWriter
}

func TestServer(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TS {
		mockWriter := newMockWriter()
		close(mockWriter.WriteOutput.Err)

		addr, err := server.Start("127.0.0.1:0", mockWriter)
		Expect(t, err == nil).To(BeTrue())

		return TS{
			T:          t,
			client:     fetchClient(addr),
			mockWriter: mockWriter,
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
}

func fetchClient(addr string) pb.DataNodeClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return pb.NewDataNodeClient(conn)
}
