package config

import (
	"log"

	"github.com/bradylove/envstruct"
)

type Config struct {
	Addr                 string   `env:"ADDR,required"`
	IntraAddr            string   `env:"INTRA_ADDR,required"`
	TalariaNodeAddr      string   `env:"TALARIA_NODE_ADDR,required"`
	TalariaSchedulerAddr string   `env:"TALARIA_SCHEDULER_ADDR,required"`
	TalariaNodeList      []string `env:"TALARIA_NODE_LIST,required"`
	IntraAnalystList     []string `env:"INTRA_ANALYST_LIST,required"`
	PprofAddr            string   `env:"PPROF_ADDR"`

	ToAnalyst map[string]string
}

func Load() *Config {
	conf := Config{
		PprofAddr: "localhost:0",
	}

	if err := envstruct.Load(&conf); err != nil {
		log.Fatalf("Invalid config: %s", err)
	}

	if len(conf.TalariaNodeList) != len(conf.IntraAnalystList) {
		log.Fatalf("List lengths of TALARIA_NODE_LIST and INTRA_ANALYST_LIST must match")
	}

	conf.ToAnalyst = make(map[string]string)
	for i := range conf.IntraAnalystList {
		conf.ToAnalyst[conf.TalariaNodeList[i]] = conf.IntraAnalystList[i]
	}

	return &conf
}
