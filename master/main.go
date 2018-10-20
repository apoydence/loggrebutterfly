package main

import (
	"log"
	"net/http"

	"github.com/poy/loggrebutterfly/master/internal/config"
	"github.com/poy/loggrebutterfly/master/internal/filesystem"
	"github.com/poy/loggrebutterfly/master/internal/rangemetrics"
	"github.com/poy/loggrebutterfly/master/internal/server"
	"github.com/poy/petasos/maintainer"

	_ "net/http/pprof"
)

func main() {
	log.Println("Starting master...")
	defer log.Println("Closing master.")

	conf := config.Load()

	metricsReader := rangemetrics.New(conf.DataNodeAddrs)
	fs := filesystem.New(conf.TalariaBufferSize, conf.SchedulerAddr, conf.TalariaNodeConverter)

	maintainer.StartBalancer(metricsReader, fs,
		maintainer.WithMinCount(conf.MinRoutes),
		maintainer.WithMaxCount(conf.MaxRoutes),
		maintainer.WithBalancerInterval(conf.BalancerInterval),
	)

	maintainer.StartFiller(metricsReader, fs,
		maintainer.WithFillerInterval(conf.FillerInterval),
		maintainer.WithFillerMinCount(conf.MinRoutes),
	)

	log.Printf("Starting server on %s", conf.Addr)
	addr, err := server.Start(conf.Addr, conf.AnalystAddrs, fs)
	if err != nil {
		log.Fatal("Unable to start server: %s", err)
	}
	log.Printf("Started server on %s", addr)

	log.Printf("Starting pprof on %s", conf.PprofAddr)
	log.Println(http.ListenAndServe(conf.PprofAddr, nil))
}
