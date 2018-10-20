package main

import (
	"log"
	"net/http"

	"github.com/poy/loggrebutterfly/datanode/internal/config"
	"github.com/poy/loggrebutterfly/datanode/internal/filesystem"
	"github.com/poy/loggrebutterfly/datanode/internal/hasher"
	"github.com/poy/loggrebutterfly/datanode/internal/server"
	"github.com/poy/loggrebutterfly/datanode/internal/server/intra"
	"github.com/poy/petasos/router"

	_ "net/http/pprof"
)

func main() {
	log.Print("Starting data node...")
	defer log.Print("Data node is closing.")

	conf := config.Load()

	fs := filesystem.New(conf.NodeAddr)
	hasher := hasher.New()
	counter := router.NewCounter()
	routerFetcher := server.NewRouterFetcher(fs, hasher, counter)

	log.Printf("Starting server on %s...", conf.Addr)
	addr, err := server.Start(conf.Addr, routerFetcher, fs)
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
	log.Printf("Started server on %s.", addr)

	log.Printf("Starting intra server on %s...", conf.IntraAddr)
	intraAddr, err := intra.Start(conf.IntraAddr, counter)
	if err != nil {
		log.Fatalf("Failed to start intra server: %s", err)
	}
	log.Printf("Started intra server on %s.", intraAddr)

	log.Printf("Starting pprof on %s", conf.PprofAddr)
	log.Println(http.ListenAndServe(conf.PprofAddr, nil))
}
