package main

import (
	"flag"

	"github.com/grokking-engineering/grokking-blog/gserver"
	"github.com/grokking-engineering/grokking-blog/utils/load-config"
	"github.com/grokking-engineering/grokking-blog/utils/logs"
)

var (
	flConfigFile = flag.String("config-file", "config-default.json", "Load config from file")

	l = logs.New("grokking-server")
)

func main() {
	flag.Parse()

	var cfg gserver.Config
	err := loadConfig.FromFileAndEnv(&cfg, *flConfigFile)
	if err != nil {
		l.WithError(err).Fatal("Loading config")
	}

	l.Fatal(gserver.Start(cfg))
}
