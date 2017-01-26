package config

import (
	"log"
	"time"

	"github.com/bradylove/envstruct"
)

type Config struct {
	Addr             string        `env:"ADDR,required"`
	SchedulerAddr    string        `env:"SCHEDULER_ADDR,required"`
	RouterAddrs      []string      `env:"ROUTER_ADDRS,required"`
	MaxRoutes        uint64        `env:"MAX_ROUTES"`
	MinRoutes        uint64        `env:"MIN_ROUTES"`
	BalancerInterval time.Duration `env:"BALANCER_INTERVAL"`
	FillerInterval   time.Duration `env:"FILLER_INTERVAL"`
}

func Load() Config {
	conf := Config{
		MaxRoutes: 10,
		MinRoutes: 4,
	}
	if err := envstruct.Load(&conf); err != nil {
		log.Fatalf("Unable to load config: %s", err)
	}

	return conf
}
