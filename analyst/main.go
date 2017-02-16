package main

import (
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/apoydence/loggrebutterfly/analyst/internal/algorithms"
	"github.com/apoydence/loggrebutterfly/analyst/internal/algorithms/mappers"
	"github.com/apoydence/loggrebutterfly/analyst/internal/algorithms/reducers"
	"github.com/apoydence/loggrebutterfly/analyst/internal/config"
	"github.com/apoydence/loggrebutterfly/analyst/internal/filesystem"
	"github.com/apoydence/loggrebutterfly/analyst/internal/network"
	"github.com/apoydence/loggrebutterfly/analyst/internal/network/intra"
	"github.com/apoydence/loggrebutterfly/analyst/internal/network/server"
	apiintra "github.com/apoydence/loggrebutterfly/api/intra"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
	"github.com/apoydence/mapreduce"
	"github.com/apoydence/talaria/api/v1"
)

func main() {
	log.Printf("Starting analyst...")
	defer log.Printf("Closing analyst.")

	conf := config.Load()

	nodeClient := setupTalariaNodeClient(conf.TalariaNodeAddr)
	schedClient := setupTalariaSchedulerClient(conf.TalariaSchedulerAddr)

	algFetcher := setupAlgorithmFetcher()
	hasher := filesystem.NewHasher()
	filter := filesystem.NewRouteFilter(hasher)
	fs := filesystem.New(filter, schedClient, nodeClient, conf.ToAnalyst)
	network := network.New()

	mr := mapreduce.New(fs, network, algFetcher)
	exec := mapreduce.NewExecutor(algFetcher, fs)

	go startIntraServer(intra.New(exec), conf.IntraAddr)
	go startServer(server.New(mr), conf.Addr)

	log.Printf("Starting pprof on %s.", conf.PprofAddr)
	log.Println(http.ListenAndServe(conf.PprofAddr, nil))
}

func setupAlgorithmFetcher() *algorithms.Fetcher {
	return algorithms.NewFetcher(map[string]algorithms.AlgBuilder{
		"timerange": algorithms.AlgBuilder(func(info *v1.AggregateInfo) mapreduce.Algorithm {
			return mapreduce.Algorithm{
				Mapper:  mappers.NewTimeRange(info),
				Reducer: reducers.NewFirst(),
			}
		}),
		"aggregation": algorithms.AlgBuilder(func(info *v1.AggregateInfo) mapreduce.Algorithm {
			return mapreduce.Algorithm{
				Mapper:  mappers.NewAggregation(info),
				Reducer: reducers.NewSumF(),
			}
		}),
	})
}

func setupTalariaNodeClient(addr string) talaria.NodeClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to talaria: %s", err)
	}
	return talaria.NewNodeClient(conn)
}

func setupTalariaSchedulerClient(addr string) talaria.SchedulerClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to talaria: %s", err)
	}
	return talaria.NewSchedulerClient(conn)
}

func startIntraServer(server *intra.Server, addr string) {
	log.Printf("Starting intra server (addr=%s)...", addr)

	lis, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	apiintra.RegisterAnalystServer(s, server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve intra: %s", err)
	}
}

func startServer(server *server.Server, addr string) {
	log.Printf("Starting server (addr=%s)...", addr)

	lis, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	v1.RegisterAnalystServer(s, server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve intra: %s", err)
	}
}
