package networkreader_test

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"

	"github.com/apoydence/eachers/testhelpers"
	"github.com/apoydence/loggrebutterfly/internal/pb/intra"
	"github.com/apoydence/loggrebutterfly/master/internal/rangemetrics/networkreader"
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

type TN struct {
	*testing.T

	reader *networkreader.NetworkReader

	routerAddrs       []string
	mockRouterServers []*mockRouterServer
}

func TestNetworkReader(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TN {
		var routerAddrs []string
		var mockRouterServers []*mockRouterServer
		for i := 0; i < 3; i++ {
			addr, m := startMockRouter()
			routerAddrs = append(routerAddrs, addr)
			mockRouterServers = append(mockRouterServers, m)
		}

		return TN{
			T:      t,
			reader: networkreader.New(),

			routerAddrs:       routerAddrs,
			mockRouterServers: mockRouterServers,
		}
	})

	o.Group("when all the routers do not return an error", func() {
		o.BeforeEach(func(t TN) TN {
			for _, m := range t.mockRouterServers {
				testhelpers.AlwaysReturn(m.ReadMetricsOutput.Ret0, &intra.ReadMetricsResponse{
					WriteCount: 5,
					ErrCount:   3,
				})
				close(m.ReadMetricsOutput.Ret1)
			}
			return t
		})

		o.Spec("it aggregates data from each router", func(t TN) {
			metric, err := t.reader.ReadMetrics(t.routerAddrs[1], "some-file")
			Expect(t, err == nil).To(BeTrue())
			Expect(t, metric.WriteCount).To(Equal(uint64(5)))
			Expect(t, metric.ErrCount).To(Equal(uint64(3)))
		})
	})
}

func startMockRouter() (string, *mockRouterServer) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	mockRouterServer := newMockRouterServer()
	s := grpc.NewServer()
	intra.RegisterRouterServer(s, mockRouterServer)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return lis.Addr().String(), mockRouterServer
}