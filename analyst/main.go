package main

import (
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/poy/loggrebutterfly/analyst/internal/algorithms"
	"github.com/poy/loggrebutterfly/analyst/internal/algorithms/mappers"
	"github.com/poy/loggrebutterfly/analyst/internal/algorithms/reducers"
	"github.com/poy/loggrebutterfly/analyst/internal/config"
	"github.com/poy/loggrebutterfly/analyst/internal/filesystem"
	"github.com/poy/loggrebutterfly/analyst/internal/network"
	"github.com/poy/loggrebutterfly/analyst/internal/network/intra"
	"github.com/poy/loggrebutterfly/analyst/internal/network/server"
	apiintra "github.com/poy/loggrebutterfly/api/intra"
	v1 "github.com/poy/loggrebutterfly/api/v1"
	"github.com/poy/mapreduce"
	"github.com/poy/talaria/api/v1"
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
		"timerange": algorithms.AlgBuilder(func(info *v1.AggregateInfo) (mapreduce.Algorithm, error) {
			filter, err := mappers.NewFilter(info)
			if err != nil {
				return mapreduce.Algorithm{}, err
			}
			return mapreduce.Algorithm{
				Mapper:  mappers.NewQuery(filter),
				Reducer: reducers.NewFirst(),
			}, nil
		}),
		"aggregation": algorithms.AlgBuilder(func(info *v1.AggregateInfo) (mapreduce.Algorithm, error) {
			filter, err := mappers.NewFilter(info)
			if err != nil {
				return mapreduce.Algorithm{}, err
			}
			agg, err := mappers.NewAggregation(info, filter)
			if err != nil {
				return mapreduce.Algorithm{}, err
			}

			return mapreduce.Algorithm{
				Mapper:  agg,
				Reducer: reducers.NewSumF(),
			}, nil
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
