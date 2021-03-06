package end2end_test

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"sort"
	"testing"
	"time"

	"github.com/poy/eachers/testhelpers"
	v2 "github.com/poy/loggrebutterfly/api/loggregator/v2"
	loggrebutterfly "github.com/poy/loggrebutterfly/api/v1"
	"github.com/poy/loggrebutterfly/internal/end2end"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
	talaria "github.com/poy/talaria/api/v1"
	"github.com/golang/protobuf/proto"
	"github.com/onsi/gomega/gexec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
		grpclog.SetLogger(log.New(ioutil.Discard, "", log.LstdFlags))
	}

	os.Exit(m.Run())
}

type TA struct {
	*testing.T
	client           loggrebutterfly.AnalystClient
	analystPort      int
	analystIntraPort int

	nodeAddr, schedAddr string
	mockNode            *mockNodeServer
	mockSched           *mockSchedulerServer
	ps                  []*os.Process
}

func TestAnalystQuery(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TA {
		ta := TA{
			T: t,
		}

		setup(&ta)
		go serveUpData(ta)
		go serveScheduler(ta)
		ta.client = startClient(fmt.Sprintf("localhost:%d", ta.analystPort))

		return ta
	})

	o.AfterEach(func(t TA) {
		for _, p := range t.ps {
			p.Kill()
		}
	})

	o.Spec("it returns only the envelopes from the right time range", func(t TA) {
		var results *loggrebutterfly.QueryResponse
		f := func() bool {
			var err error
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			results, err = t.client.Query(ctx, &loggrebutterfly.QueryInfo{
				Filter: &loggrebutterfly.AnalystFilter{
					SourceId: "some-id",
					TimeRange: &loggrebutterfly.TimeRange{
						Start: 10,
						End:   20,
					},
				},
			})
			return err == nil
		}

		Expect(t, f).To(ViaPollingMatcher{
			Duration: 5 * time.Second,
			Matcher:  BeTrue(),
		})

		sort.Sort(envelopes(results.Envelopes))

		Expect(t, results.Envelopes).To(HaveLen(10))
		for i, e := range results.Envelopes {
			Expect(t, e.SourceId).To(Equal("some-id"))
			Expect(t, e.Timestamp).To(Equal(int64(i + 10)))
		}
	})

	o.Spec("it returns only the envelopes for the correct counter name", func(t TA) {
		var results *loggrebutterfly.QueryResponse
		f := func() bool {
			var err error
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			results, err = t.client.Query(ctx, &loggrebutterfly.QueryInfo{
				Filter: &loggrebutterfly.AnalystFilter{
					SourceId: "some-id",
					TimeRange: &loggrebutterfly.TimeRange{
						Start: 0,
						End:   9223372036854775807,
					},
					Envelopes: &loggrebutterfly.AnalystFilter_Counter{
						Counter: &loggrebutterfly.CounterFilter{
							Name: "some-name-0",
						},
					},
				},
			})
			return err == nil
		}

		Expect(t, f).To(ViaPollingMatcher{
			Duration: 5 * time.Second,
			Matcher:  BeTrue(),
		})

		sort.Sort(envelopes(results.Envelopes))

		Expect(t, results.Envelopes).To(HaveLen(50))
		for i, e := range results.Envelopes {
			Expect(t, e.SourceId).To(Equal("some-id"))
			Expect(t, e.Timestamp).To(Equal(int64(i * 2)))
		}
	})
}

func TestAnalystAggregate(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TA {
		ta := TA{
			T: t,
		}

		setup(&ta)
		go serveUpData(ta)
		go serveScheduler(ta)
		ta.client = startClient(fmt.Sprintf("localhost:%d", ta.analystPort))

		return ta
	})

	o.AfterEach(func(t TA) {
		for _, p := range t.ps {
			p.Kill()
		}
	})

	o.Spec("it aggregates data into the requested buckets", func(t TA) {
		var results *loggrebutterfly.AggregateResponse
		f := func() bool {
			var err error
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			results, err = t.client.Aggregate(ctx, &loggrebutterfly.AggregateInfo{
				BucketWidthNs: 5,
				Query: &loggrebutterfly.QueryInfo{
					Filter: &loggrebutterfly.AnalystFilter{
						SourceId: "some-id",
						TimeRange: &loggrebutterfly.TimeRange{
							Start: 10,
							End:   20,
						},
						Envelopes: &loggrebutterfly.AnalystFilter_Counter{
							Counter: &loggrebutterfly.CounterFilter{},
						},
					},
				},
			})
			return err == nil
		}

		Expect(t, f).To(ViaPollingMatcher{
			Duration: 5 * time.Second,
			Matcher:  BeTrue(),
		})

		Expect(t, results.Results).To(HaveLen(2))
		Expect(t, results.Results[10]).To(Equal(float64(99 * 5)))
		Expect(t, results.Results[15]).To(Equal(float64(99 * 5)))
	})

	o.Spec("it aggregates data into the requested buckets for the filtered name", func(t TA) {
		var results *loggrebutterfly.AggregateResponse
		f := func() bool {
			var err error
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			results, err = t.client.Aggregate(ctx, &loggrebutterfly.AggregateInfo{
				BucketWidthNs: 5,
				Query: &loggrebutterfly.QueryInfo{
					Filter: &loggrebutterfly.AnalystFilter{
						SourceId: "some-id",
						TimeRange: &loggrebutterfly.TimeRange{
							Start: 10,
							End:   20,
						},
						Envelopes: &loggrebutterfly.AnalystFilter_Counter{
							Counter: &loggrebutterfly.CounterFilter{
								Name: "some-name-0",
							},
						},
					},
				},
			})
			return err == nil
		}

		Expect(t, f).To(ViaPollingMatcher{
			Duration: 5 * time.Second,
			Matcher:  BeTrue(),
		})

		Expect(t, results.Results).To(HaveLen(2))
		Expect(t, results.Results[10]).To(Equal(float64(99 * 3)))
		Expect(t, results.Results[15]).To(Equal(float64(99 * 2)))
	})
}

func serveUpData(t TA) {
	for {
		rx := <-t.mockNode.ReadInput.Arg1
		go func(rx talaria.Node_ReadServer) {
			for i := 0; i < 100; i++ {
				e := &v2.Envelope{
					SourceId:  "some-id",
					Timestamp: int64(i),
					Message: &v2.Envelope_Counter{
						Counter: &v2.Counter{
							Name:  fmt.Sprintf("some-name-%d", i%2),
							Value: &v2.Counter_Total{99},
						},
					},
				}
				data, err := proto.Marshal(e)
				if err != nil {
					panic(err)
				}
				err = rx.Send(&talaria.ReadDataPacket{Message: data})
				if err != nil {
					return
				}
			}

			t.mockNode.ReadOutput.Ret0 <- io.EOF
		}(rx)
	}
}

func serveScheduler(t TA) {
	close(t.mockSched.ListClusterInfoOutput.Ret1)
	testhelpers.AlwaysReturn(t.mockSched.ListClusterInfoOutput.Ret0, &talaria.ListResponse{
		Info: []*talaria.ClusterInfo{
			{
				Name: `{"Low":0,"High":18446744073709551615}`,
				Nodes: []*talaria.NodeInfo{
					{
						URI: t.nodeAddr,
					},
				},
			},
		},
	})
}

func startClient(addr string) loggrebutterfly.AnalystClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return loggrebutterfly.NewAnalystClient(conn)
}

func setup(t *TA) {
	t.nodeAddr, t.mockNode = startMockNode()
	t.schedAddr, t.mockSched = startMockSched()

	t.analystPort = end2end.AvailablePort()
	t.analystIntraPort = end2end.AvailablePort()
	analystPs := startAnalyst(t.analystPort, t.analystIntraPort, t.nodeAddr, t.schedAddr)
	t.ps = append(t.ps, analystPs)
}

func startAnalyst(port, intraPort int, dataNodeAddr, schedAddr string) *os.Process {
	log.Printf("Starting analyst on %d...", port)
	defer log.Printf("Done starting analyst on %d.", port)

	path, err := gexec.Build("github.com/poy/loggrebutterfly/analyst")
	if err != nil {
		panic(err)
	}

	command := exec.Command(path)
	command.Env = []string{
		fmt.Sprintf("ADDR=127.0.0.1:%d", port),
		fmt.Sprintf("INTRA_ADDR=127.0.0.1:%d", intraPort),
		fmt.Sprintf("TALARIA_NODE_ADDR=%s", dataNodeAddr),
		fmt.Sprintf("TALARIA_SCHEDULER_ADDR=%s", schedAddr),
		fmt.Sprintf("TALARIA_NODE_LIST=%s", dataNodeAddr),
		fmt.Sprintf("INTRA_ANALYST_LIST=127.0.0.1:%d", intraPort),
	}

	if testing.Verbose() {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	if err = command.Start(); err != nil {
		panic(err)
	}

	return command.Process
}

func startMockNode() (string, *mockNodeServer) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	mockNodeServer := newMockNodeServer()
	s := grpc.NewServer()
	talaria.RegisterNodeServer(s, mockNodeServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	log.Printf("Starting talaria node on %s", lis.Addr().String())
	return lis.Addr().String(), mockNodeServer
}

func startMockSched() (string, *mockSchedulerServer) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	mockSchedServer := newMockSchedulerServer()
	s := grpc.NewServer()
	talaria.RegisterSchedulerServer(s, mockSchedServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	log.Printf("Starting talaria sched on %s", lis.Addr().String())
	return lis.Addr().String(), mockSchedServer
}

type envelopes []*v2.Envelope

func (e envelopes) Len() int {
	return len(e)
}

func (e envelopes) Less(i, j int) bool {
	return e[i].Timestamp < e[j].Timestamp
}

func (e envelopes) Swap(i, j int) {
	tmp := e[i]
	e[i] = e[j]
	e[j] = tmp
}
