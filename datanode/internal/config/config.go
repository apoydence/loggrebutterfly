package config

import (
	"log"

	"github.com/bradylove/envstruct"
)

type Config struct {
	Addr      string `env:"ADDR,required"`
	IntraAddr string `env:"INTRA_ADDR,required"`
	NodeAddr  string `env:"NODE_ADDR,required"`
	PprofAddr string `env:"PPROF_ADDR"`
}

func Load() Config {
	conf := Config{
		PprofAddr: "localhost:0",
	}
	if err := envstruct.Load(&conf); err != nil {
		log.Fatalf("Unable to load config: %s", err)
	}

	return conf
}
