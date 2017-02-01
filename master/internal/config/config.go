package config

import (
	"log"
	"time"

	"github.com/bradylove/envstruct"
)

type Config struct {
	Addr      string `env:"ADDR,required"`
	PprofAddr string `env:"PPROF_ADDR"`

	SchedulerAddr        string   `env:"SCHEDULER_ADDR,required"`
	DataNodeAddrs        []string `env:"DATA_NODE_ADDRS,required"`
	DataNodeExtAddrs     []string `env:"DATA_NODE_EXTERNAL_ADDRS,required"`
	TalariaNodeAddrs     []string `env:"TALARIA_NODE_ADDRS,required"`
	TalariaNodeConverter map[string]string

	MaxRoutes        uint64        `env:"MAX_ROUTES"`
	MinRoutes        uint64        `env:"MIN_ROUTES"`
	BalancerInterval time.Duration `env:"BALANCER_INTERVAL"`
	FillerInterval   time.Duration `env:"FILLER_INTERVAL"`
}

func Load() Config {
	conf := Config{
		MaxRoutes:        10,
		MinRoutes:        4,
		BalancerInterval: 5 * time.Second,
		FillerInterval:   time.Second,
		PprofAddr:        "localhost:0",
	}
	if err := envstruct.Load(&conf); err != nil {
		log.Fatalf("Unable to load config: %s", err)
	}

	if len(conf.DataNodeExtAddrs) != len(conf.TalariaNodeAddrs) {
		log.Fatalf("DATA_NODE_EXTERNAL_ADDRS (%d) and TALARIA_NODE_ADDRS (%d) must have same count", len(conf.DataNodeExtAddrs), len(conf.TalariaNodeAddrs))
	}

	conf.TalariaNodeConverter = make(map[string]string)
	for i := range conf.DataNodeAddrs {
		conf.TalariaNodeConverter[conf.TalariaNodeAddrs[i]] = conf.DataNodeExtAddrs[i]
	}

	return conf
}
