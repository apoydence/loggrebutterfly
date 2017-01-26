package main

import (
	"log"

	"github.com/apoydence/loggrebutterfly/master/internal/config"
	"github.com/apoydence/loggrebutterfly/master/internal/filesystem"
	"github.com/apoydence/loggrebutterfly/master/internal/rangemetrics"
	"github.com/apoydence/petasos/maintainer"
)

func main() {
	log.Println("Starting master...")
	defer log.Println("Closing master.")

	conf := config.Load()

	metricsReader := rangemetrics.New(conf.RouterAddrs)
	fs := filesystem.New(conf.SchedulerAddr)

	maintainer.StartBalancer(metricsReader, fs,
		maintainer.WithMinCount(conf.MinRoutes),
		maintainer.WithMaxCount(conf.MaxRoutes),
		maintainer.WithBalancerInterval(conf.BalancerInterval),
	)

	maintainer.StartFiller(metricsReader, fs,
		maintainer.WithFillerInterval(conf.FillerInterval),
	)

	var c chan struct{}
	<-c
}
