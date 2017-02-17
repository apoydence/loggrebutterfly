package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	v2 "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	pb "github.com/apoydence/loggrebutterfly/api/v1"
	"github.com/apoydence/loggrebutterfly/client"
	"github.com/apoydence/petasos/router"
	"github.com/golang/protobuf/jsonpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	verbose = flag.Bool("verbose", false, "Verbose mode")

	masterAddr      = flag.String("master", "", "The address for the master")
	sourceId        = flag.String("source-id", "", "The source-id to interact with")
	writePacketSize = flag.Uint("packetSize", 1024, "The size of each write packet")

	analyticsStart          = flag.Int64("analytics-start", -1, "The start parameter for a query")
	analyticsEnd            = flag.Int64("analytics-end", -1, "The end parameter for a query")
	analyticsCounterName    = flag.String("counter-name", "", "The name of the counters to aggregate")
	analyticsBucketDuration = flag.String("bucket-duration", "", "The duration width of each bucket")

	showHash   = flag.Bool("show-hash", false, "Show the hash of a source-id")
	tail       = flag.Bool("tail", false, "The tail the source-id")
	writeData  = flag.Bool("write", false, "Write data from STDIN")
	listRoutes = flag.Bool("list", false, "List the routes")

	query = flag.Bool("query", false, "Query the data")

	showCounter = flag.Bool("show-counter-example", false, "Show an example (JSON) counter envelope")
)

func main() {
	flag.Parse()

	if !*verbose {
		grpclog.SetLogger(log.New(ioutil.Discard, "", log.LstdFlags))
	}

	if *showCounter {
		e := &v2.Envelope{
			SourceId:  "some-id",
			Timestamp: 99,
			Message: &v2.Envelope_Counter{
				Counter: &v2.Counter{
					Name: "counter-name",
					Value: &v2.Counter_Total{
						Total: 101,
					},
				},
			},
		}

		data, _ := new(jsonpb.Marshaler).MarshalToString(e)
		fmt.Println(data)
		return
	}

	if *showHash {
		if *sourceId == "" {
			log.Fatal("You must provide source-id")
		}

		f := fnv.New64a()
		f.Write([]byte(*sourceId))

		fmt.Println(f.Sum64())
		return
	}

	if *masterAddr == "" {
		log.Fatal("You must provide a master address")
	}

	if *tail {
		tailCommand(client.New(*masterAddr))
		return
	}

	if *writeData {
		writeDataCommand(client.New(*masterAddr))
		return
	}

	if *listRoutes {
		list()
		return
	}

	if *query {
		queryData()
		return
	}

	onlyOneCommandUsage()
}

func setupMasterClient() pb.MasterClient {
	conn, err := grpc.Dial(*masterAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to scheduler: %s", err)
	}
	return pb.NewMasterClient(conn)
}

func setupDataNodeClient(addr string) pb.DataNodeClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to node: %s", err)
	}
	return pb.NewDataNodeClient(conn)
}

func setupAnalystClient(addr string) pb.AnalystClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to node: %s", err)
	}
	return pb.NewAnalystClient(conn)
}

func tailCommand(client *client.Client) {
	if *writeData || *listRoutes || *query {
		onlyOneCommandUsage()
	}

	if *sourceId == "" {
		log.Fatal("You must provide a source-id")
	}

	rx := client.ReadFrom(*sourceId)

	var i int
	for {
		envelope, err := rx()
		if err == io.EOF {
			time.Sleep(250 * time.Millisecond)
			continue
		}

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("(%d) %+v\n", i, envelope)
		i++
	}
}

func writeDataCommand(client *client.Client) {
	if *tail || *listRoutes || *query {
		onlyOneCommandUsage()
	}

	defer func() {
		// Give the buffer time to clear
		time.Sleep(250 * time.Millisecond)
	}()

	var count int
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		var e v2.Envelope
		if err := jsonpb.UnmarshalString(string(scanner.Bytes()), &e); err != nil {
			log.Fatalf("unable to parse (via json) (%s) to v2.Envelope: %s", string(scanner.Bytes()), err)
		}

		for i := 0; i < 10; i++ {
			var err error
			if err = client.Write(&e); err == nil {
				log.Printf("Successfully wrote data: %v\n", e)
				break
			}
			log.Println("Error writing", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		count++
	}

}

func list() {
	if *writeData || *tail || *query {
		onlyOneCommandUsage()
	}

	masterClient := setupMasterClient()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := masterClient.Routes(ctx, new(pb.RoutesInfo))
	if err != nil {
		log.Fatal(err)
	}

	sort.Sort(routes(resp.Routes))

	for i, r := range resp.Routes {
		fmt.Printf("Route %d: %s (Leader=%s)\n", i, r.Name, r.Leader)
	}
	fmt.Printf("Listed %d routes\n", len(resp.Routes))
}

func queryData() {
	if *writeData || *tail || *listRoutes {
		onlyOneCommandUsage()
	}

	if *analyticsStart == -1 || *analyticsEnd == -1 || *sourceId == "" {
		log.Fatal("analytics-start, analytics-end and source-id are required")
	}

	masterClient := setupMasterClient()
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := masterClient.Analysts(ctx, new(pb.AnalystsInfo))
	if err != nil {
		log.Fatal(err)
	}

	if len(resp.Analysts) == 0 {
		log.Fatal("Master did not report any analyst nodes")
	}

	analystAddr := resp.Analysts[rand.Intn(len(resp.Analysts))]
	log.Printf("Using analyst %s", analystAddr.Addr)
	analystClient := setupAnalystClient(analystAddr.Addr)

	if *analyticsCounterName == "" {
		ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
		queryResp, err := analystClient.Query(ctx, &pb.QueryInfo{
			SourceId: *sourceId,
			TimeRange: &pb.TimeRange{
				Start: *analyticsStart,
				End:   *analyticsEnd,
			},
		})
		if err != nil {
			log.Fatalf("Executing query failed: %s", err)
		}

		sort.Sort(envelopes(queryResp.Envelopes))

		for i, e := range queryResp.Envelopes {
			fmt.Printf("Envelope %d: %+v\n", i, e)
		}
		fmt.Printf("Printed %d results", len(queryResp.Envelopes))
		return
	}

	if *analyticsBucketDuration == "" {
		log.Fatal("bucket-duration must be provided")
	}

	d, err := time.ParseDuration(*analyticsBucketDuration)
	if err != nil {
		log.Fatalf("Failed to parse bucket-duration:", err)
	}

	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	results, err := analystClient.Aggregate(ctx, &pb.AggregateInfo{
		BucketWidthNs: int64(d),
		Query: &pb.QueryInfo{
			SourceId: *sourceId,
			TimeRange: &pb.TimeRange{
				Start: *analyticsStart,
				End:   *analyticsEnd,
			},
		},
		Aggregation: &pb.AggregateInfo_Counter{
			Counter: &pb.CounterAggregation{
				Name: *analyticsCounterName,
			},
		},
	})
	if err != nil {
		log.Fatalf("Executing query failed: %s", err)
	}

	var fr []floatResult
	for k, v := range results.Results {
		fr = append(fr, floatResult{t: k, f: v})
	}

	sort.Sort(floatResults(fr))
	for _, v := range fr {
		fmt.Printf("%d -> %v\n", v.t, v.f)
	}
}

func onlyOneCommandUsage() {
	log.Fatal("Use only one tail, write, list or query")
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

func mapToEnvelope(m map[string]interface{}) *v2.Envelope {
	var e v2.Envelope
	e.SourceId = m["source_id"].(string)
	e.Timestamp = m["source_id"].(int64)

	return &e
}

type floatResult struct {
	t int64
	f float64
}
type floatResults []floatResult

func (r floatResults) Len() int {
	return len(r)
}

func (r floatResults) Less(i, j int) bool {
	return r[i].t < r[j].t
}

func (r floatResults) Swap(i, j int) {
	tmp := r[i]
	r[i] = r[j]
	r[j] = tmp
}

type routes []*pb.RouteInfo

func (r routes) Len() int {
	return len(r)
}

func (r routes) Less(i, j int) bool {
	var rn1, rn2 router.RangeName
	json.Unmarshal([]byte(r[i].Name), &rn1)
	json.Unmarshal([]byte(r[j].Name), &rn2)
	return rn1.Term < rn2.Term
}

func (r routes) Swap(i, j int) {
	tmp := r[i]
	r[i] = r[j]
	r[j] = tmp
}
