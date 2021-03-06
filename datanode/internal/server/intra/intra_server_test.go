//go:generate hel

package intra_test

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"google.golang.org/grpc"

	"github.com/poy/loggrebutterfly/datanode/internal/server/intra"
	pb "github.com/poy/loggrebutterfly/api/intra"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
	"github.com/poy/petasos/router"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	os.Exit(m.Run())
}

type TI struct {
	*testing.T
	client            pb.DataNodeClient
	mockMetricsReader *mockMetricsReader
}

func TestIntra(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TI {
		mockMetricsReader := newMockMetricsReader()

		addr, err := intra.Start("127.0.0.1:0", mockMetricsReader)
		Expect(t, err == nil).To(BeTrue())

		return TI{
			T:                 t,
			client:            fetchClient(addr),
			mockMetricsReader: mockMetricsReader,
		}
	})

	o.Spec("reports the metrics from the metrics reader", func(t TI) {
		t.mockMetricsReader.MetricsOutput.Metric <- router.Metric{
			WriteCount: 99,
			ErrCount:   101,
		}

		resp, err := t.client.ReadMetrics(context.Background(), &pb.ReadMetricsInfo{
			File: `{"Term":99}`,
		})
		Expect(t, err == nil).To(BeTrue())
		Expect(t, resp.WriteCount).To(Equal(uint64(99)))
		Expect(t, resp.ErrCount).To(Equal(uint64(101)))
	})
}

func fetchClient(addr string) pb.DataNodeClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return pb.NewDataNodeClient(conn)
}
