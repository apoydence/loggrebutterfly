package end2end_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/apoydence/eachers/testhelpers"
	"github.com/apoydence/loggrebutterfly/internal/end2end"
	"github.com/apoydence/loggrebutterfly/pb/intra"
	"github.com/apoydence/onpar"
	"github.com/apoydence/petasos/router"
	"github.com/apoydence/talaria/pb"
	"github.com/onsi/gomega/gexec"

	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
)

var (
	masterPort    int
	schedAddr     string
	mockScheduler *mockSchedulerServer
	mockDataNodes []*mockDataNodeServer
)

func TestMain(m *testing.M) {
	flag.Parse()

	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	ps := setup()

	var status int
	func() {
		defer func() {
			for _, p := range ps {
				p.Kill()
			}
		}()
		status = m.Run()
	}()

	os.Exit(status)
}

type TM struct {
	*testing.T
}

func TestMaster(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	var mu sync.Mutex
	var createResults []string

	go func() {
		for {
			x := <-mockScheduler.CreateInput.Arg1
			mu.Lock()
			createResults = append(createResults, x.Name)
			mu.Unlock()
		}
	}()

	go func() {
		for {
			<-mockScheduler.ListClusterInfoCalled
			var info []*pb.ClusterInfo
			mu.Lock()
			for _, name := range createResults {
				info = append(info, &pb.ClusterInfo{Name: name})
			}
			mu.Unlock()

			mockScheduler.ListClusterInfoOutput.Ret0 <- &pb.ListResponse{Info: info}
			mockScheduler.ListClusterInfoOutput.Ret1 <- nil
		}
	}()

	testhelpers.AlwaysReturn(mockScheduler.CreateOutput.Ret0, new(pb.CreateResponse))
	close(mockScheduler.CreateOutput.Ret1)

	testhelpers.AlwaysReturn(mockScheduler.ReadOnlyOutput.Ret0, new(pb.ReadOnlyResponse))
	close(mockScheduler.ReadOnlyOutput.Ret1)

	for _, m := range mockDataNodes {
		testhelpers.AlwaysReturn(m.ReadMetricsOutput.Ret0, new(intra.ReadMetricsResponse))
		close(m.ReadMetricsOutput.Ret1)
	}

	o.BeforeEach(func(t *testing.T) TM {
		return TM{
			T: t,
		}
	})

	o.Spec("it creates 4 buffers", func(t TM) {

		f := func() bool {
			return len(mockScheduler.CreateCalled) >= 4
		}

		Expect(t, f).To(ViaPollingMatcher{
			Matcher:  BeTrue(),
			Duration: 10 * time.Second,
		})
	})

	o.Spec("it fills a gap", func(t TM) {
		mockScheduler.CreateInput.Arg1 <- &pb.CreateInfo{Name: buildRangeName(0, 9223372036854775807, 100)}

		f := func() bool {
			return len(mockScheduler.CreateCalled) >= 5
		}

		Expect(t, f).To(ViaPollingMatcher{
			Matcher:  BeTrue(),
			Duration: 10 * time.Second,
		})
	})
}

func setup() []*os.Process {
	schedAddr, mockScheduler = startMockScheduler()

	var routers []string
	for i := 0; i < 3; i++ {
		addr, m := startMockDataNode()
		routers = append(routers, addr)
		mockDataNodes = append(mockDataNodes, m)
	}

	masterPort = end2end.AvailablePort()
	masterPs := startMaster(masterPort, schedAddr, routers)

	return []*os.Process{
		masterPs,
	}
}

func startMaster(port int, schedAddr string, routers []string) *os.Process {
	log.Printf("Starting master on %d...", port)
	defer log.Printf("Done starting master on %d.", port)

	path, err := gexec.Build("github.com/apoydence/loggrebutterfly/master")
	if err != nil {
		panic(err)
	}

	command := exec.Command(path)
	command.Env = []string{
		fmt.Sprintf("ADDR=127.0.0.1:%d", port),
		fmt.Sprintf("SCHEDULER_ADDR=%s", schedAddr),
		fmt.Sprintf("ROUTER_ADDRS=%s", buildDataNodeAddrs(routers)),
		"BALANCER_INTERVAL=1ms",
		"FILLER_INTERVAL=1ms",
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

func startMockScheduler() (string, *mockSchedulerServer) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	mockSchedulerServer := newMockSchedulerServer()
	s := grpc.NewServer()
	pb.RegisterSchedulerServer(s, mockSchedulerServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return lis.Addr().String(), mockSchedulerServer
}

func startMockDataNode() (string, *mockDataNodeServer) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	mockDataNodeServer := newMockDataNodeServer()
	s := grpc.NewServer()
	intra.RegisterDataNodeServer(s, mockDataNodeServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return lis.Addr().String(), mockDataNodeServer
}

func buildDataNodeAddrs(addrs []string) string {
	return strings.Join(addrs, ",")
}

func buildRangeName(low, high, term uint64) string {
	rn := router.RangeName{
		Low:  low,
		High: high,
		Term: term,
	}

	j, _ := json.Marshal(rn)
	return string(j)
}
