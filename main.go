package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/Xynnn007/DFS/config"
	"github.com/Xynnn007/DFS/network/http"
	"github.com/Xynnn007/DFS/raft"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	path := flag.String("p", "./test/peers.toml", "Path to the peers.toml")
	flag.Parse()

	cfg, err := config.New(*path)
	if err != nil {
		log.Fatalf("Config read failed: %v", err)
	}

	n := http.New(cfg.Config.Port)

	r := raft.New(cfg.Config)
	r.SetTransporter(n)
	r.Start()

	c := make(chan os.Signal)
	signal.Notify(c)
	<-c
	log.Info("Get EXIT signal, exited!")
}
