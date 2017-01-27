package main

import (
	"log"
	"net/http"

	"github.com/apoydence/loggrebutterfly/datanode/internal/config"
	"github.com/apoydence/loggrebutterfly/datanode/internal/filesystem"
	"github.com/apoydence/loggrebutterfly/datanode/internal/hasher"
	"github.com/apoydence/loggrebutterfly/datanode/internal/server"
	"github.com/apoydence/loggrebutterfly/datanode/internal/server/intra"
	"github.com/apoydence/petasos/router"
)

func main() {
	log.Print("Starting data node...")
	defer log.Print("Data node is closing.")

	conf := config.Load()

	fs := filesystem.New(conf.NodeAddr)
	hasher := hasher.New()
	router := router.New(fs, hasher)

	log.Printf("Starting server on %s...", conf.Addr)
	addr, err := server.Start(conf.Addr, router)
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
	log.Printf("Started server on %s.", addr)

	log.Printf("Starting intra server on %s...", conf.IntraAddr)
	intraAddr, err := intra.Start(conf.IntraAddr, router)
	if err != nil {
		log.Fatalf("Failed to start intra server: %s", err)
	}
	log.Printf("Started intra server on %s.", intraAddr)

	log.Printf("Starting pprof on %s", conf.PprofAddr)
	log.Println(http.ListenAndServe(conf.PprofAddr, nil))
}
