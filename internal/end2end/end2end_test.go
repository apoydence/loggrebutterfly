package end2end_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	v2 "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	pb "github.com/apoydence/loggrebutterfly/api/v1"
	"github.com/apoydence/loggrebutterfly/client"
	"github.com/apoydence/loggrebutterfly/internal/end2end"
	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
	"github.com/apoydence/petasos/router"
	"github.com/onsi/gomega/gexec"
)

var (
	masterPort   int
	schedPort    int
	analystPorts []int
)

func TestMain(m *testing.M) {
	flag.Parse()

	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
	grpclog.SetLogger(log.New(ioutil.Discard, "", log.LstdFlags))

	log.Println("Setting up...")
	ps := setup()
	log.Println("Done setting up.")

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
	client pb.MasterClient
}

func TestMaster(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TM {
		return TM{
			T:      t,
			client: fetchMasterClient(masterPort),
		}
	})

	o.Spec("it writes and reads from Talaria", func(t TM) {
		var resp *pb.RoutesResponse
		f := func() bool {
			var err error
			ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
			resp, err = t.client.Routes(ctx, new(pb.RoutesInfo))
			if err != nil {
				return false
			}

			for _, r := range resp.Routes {
				if r.Leader == "" {
					return false
				}
			}

			return len(resp.Routes) == 4
		}
		Expect(t, f).To(ViaPollingMatcher{
			Duration: 5 * time.Second,
			Matcher:  BeTrue(),
		})

		client := client.New(fmt.Sprintf("127.0.0.1:%d", masterPort))

		e := &v2.Envelope{
			SourceUuid: "some-id",
			Timestamp:  99,
		}

		reader := client.ReadFrom("some-id")

		var rxEnvelope *v2.Envelope
		f = func() bool {
			err := client.Write(e)
			if err != nil {
				return false
			}

			rxEnvelope, err = reader()
			return err == nil
		}
		Expect(t, f).To(ViaPollingMatcher{
			Duration: 3 * time.Second,
			Matcher:  BeTrue(),
		})
		Expect(t, rxEnvelope).To(Equal(e))

		analyst := fetchAnalystClient(analystPorts[0])
		var queryResp *pb.QueryResponse
		f = func() bool {
			var err error
			queryResp, err = analyst.Query(context.Background(), &pb.QueryInfo{
				SourceUuid: "some-id",
				TimeRange: &pb.TimeRange{
					Start: 99,
					End:   10000,
				},
			})

			return err == nil
		}
		Expect(t, f).To(ViaPollingMatcher{
			Duration: 3 * time.Second,
			Matcher:  BeTrue(),
		})
		Expect(t, queryResp.Envelopes).To(Not(HaveLen(0)))
	})
}

func setup() []*os.Process {
	var (
		dataNodePorts      []int
		dataNodeIntraPorts []int
		nodePorts          []int
		nodeIntraPorts     []int

		analystIntraPorts []int

		ps []*os.Process
	)

	for i := 0; i < 3; i++ {
		port, intraPort, nodePort, intraNodePort, p := startDataNode()
		dataNodePorts = append(dataNodePorts, port)
		dataNodeIntraPorts = append(dataNodeIntraPorts, intraPort)
		nodePorts = append(nodePorts, nodePort)
		nodeIntraPorts = append(nodeIntraPorts, intraNodePort)
		ps = append(ps, p...)
	}

	var masterPs []*os.Process
	masterPort, schedPort, masterPs = startMaster(dataNodeIntraPorts, dataNodePorts, nodePorts, nodeIntraPorts)
	ps = append(ps, masterPs...)

	for i := 0; i < 3; i++ {
		analystIntraPorts = append(analystIntraPorts, end2end.AvailablePort())
	}

	for i := 0; i < 3; i++ {
		port, p := startAnalyst(analystIntraPorts[i], nodePorts[i], schedPort, analystIntraPorts, nodeIntraPorts)
		analystPorts = append(analystPorts, port)
		ps = append(ps, p)
	}

	return ps
}

func fetchMasterClient(port int) pb.MasterClient {
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithInsecure())
	if err != nil {
		return nil
	}
	return pb.NewMasterClient(conn)
}

func fetchAnalystClient(port int) pb.AnalystClient {
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithInsecure())
	if err != nil {
		return nil
	}
	return pb.NewAnalystClient(conn)
}

func startMaster(routerPorts, extRouterPorts, nodePorts, nodeIntraPorts []int) (port, schedPort int, ps []*os.Process) {
	schedPort, schedPs := startTalariaScheduler(nodeIntraPorts)

	port = end2end.AvailablePort()
	log.Printf("Starting master on %d...", port)
	defer log.Printf("Done starting master on %d.", port)

	path, err := gexec.Build("github.com/apoydence/loggrebutterfly/master")
	if err != nil {
		panic(err)
	}

	command := exec.Command(path)
	command.Env = []string{
		fmt.Sprintf("ADDR=127.0.0.1:%d", port),
		fmt.Sprintf("SCHEDULER_ADDR=127.0.0.1:%d", schedPort),
		fmt.Sprintf("DATA_NODE_ADDRS=%s", buildNodeURIs(routerPorts)),
		fmt.Sprintf("DATA_NODE_EXTERNAL_ADDRS=%s", buildNodeURIs(extRouterPorts)),
		fmt.Sprintf("TALARIA_NODE_ADDRS=%s", buildNodeURIs(nodePorts)),
		"ANALYST_ADDRS=left-out",
		"BALANCER_INTERVAL=1s",
		"FILLER_INTERVAL=1s",
	}

	if testing.Verbose() {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	if err = command.Start(); err != nil {
		panic(err)
	}

	return port, schedPort, []*os.Process{command.Process, schedPs}
}

func startTalariaScheduler(nodePorts []int) (port int, ps *os.Process) {
	port = end2end.AvailablePort()
	log.Printf("Scheduler Port = %d", port)
	for i, nodePort := range nodePorts {
		log.Printf("Node Port (%d) = %d", i, nodePort)
	}

	path, err := gexec.Build("github.com/apoydence/talaria/scheduler")
	if err != nil {
		panic(err)
	}

	command := exec.Command(path)
	command.Env = []string{
		fmt.Sprintf("ADDR=localhost:%d", port),
		fmt.Sprintf("NODES=%s", buildNodeURIs(nodePorts)),
	}

	if testing.Verbose() {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	err = command.Start()
	if err != nil {
		panic(err)
	}

	return port, command.Process
}

func startDataNode() (port, intraPort, nodePort, intraNodePort int, ps []*os.Process) {
	var nodePs *os.Process
	nodePort, intraNodePort, nodePs = startTalariaNode()

	port = end2end.AvailablePort()
	intraPort = end2end.AvailablePort()

	log.Printf("Starting data node on %d (talaria=%d)...\n", port, nodePort)
	log.Printf("Starting data node on %d (talaria=%d)...", port, nodePort)
	defer log.Printf("Done starting data node on %d.", port)

	path, err := gexec.Build("github.com/apoydence/loggrebutterfly/datanode")
	if err != nil {
		panic(err)
	}

	command := exec.Command(path)
	command.Env = []string{
		fmt.Sprintf("ADDR=127.0.0.1:%d", port),
		fmt.Sprintf("INTRA_ADDR=127.0.0.1:%d", intraPort),
		fmt.Sprintf("NODE_ADDR=127.0.0.1:%d", nodePort),
	}

	if testing.Verbose() {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	if err = command.Start(); err != nil {
		panic(err)
	}

	return port, intraPort, nodePort, intraNodePort, []*os.Process{command.Process, nodePs}
}

func startAnalyst(
	intraNodePort int,
	talariaNodePort int,
	talariaSchedPort int,
	intraPorts []int,
	talariaNodePorts []int,
) (port int, ps *os.Process) {
	nodePort := end2end.AvailablePort()
	path, err := gexec.Build("github.com/apoydence/loggrebutterfly/analyst")
	if err != nil {
		panic(err)
	}
	command := exec.Command(path)
	command.Env = []string{
		fmt.Sprintf("ADDR=localhost:%d", nodePort),
		fmt.Sprintf("INTRA_ADDR=localhost:%d", intraNodePort),
		fmt.Sprintf("TALARIA_NODE_ADDR=localhost:%d", talariaNodePort),
		fmt.Sprintf("TALARIA_SCHEDULER_ADDR=localhost:%d", talariaSchedPort),
		fmt.Sprintf("TALARIA_NODE_LIST=%s", buildNodeURIs(talariaNodePorts)),
		fmt.Sprintf("INTRA_ANALYST_LIST=%s", buildNodeURIs(intraPorts)),
	}

	if testing.Verbose() {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	err = command.Start()
	if err != nil {
		panic(err)
	}

	return nodePort, command.Process
}

func startTalariaNode() (int, int, *os.Process) {
	nodePort := end2end.AvailablePort()
	intraNodePort := end2end.AvailablePort()
	path, err := gexec.Build("github.com/apoydence/talaria/node")
	if err != nil {
		panic(err)
	}
	command := exec.Command(path)
	command.Env = []string{
		fmt.Sprintf("ADDR=localhost:%d", nodePort),
		fmt.Sprintf("INTRA_ADDR=localhost:%d", intraNodePort),
	}

	err = command.Start()
	if err != nil {
		panic(err)
	}

	return nodePort, intraNodePort, command.Process
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

func buildNodeURIs(ports []int) string {
	var URIs []string
	for _, port := range ports {
		URIs = append(URIs, fmt.Sprintf("127.0.0.1:%d", port))
	}
	return strings.Join(URIs, ",")
}
