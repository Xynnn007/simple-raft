package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/Xynnn007/DFS/config"
	"github.com/Xynnn007/DFS/network/http"
	"github.com/Xynnn007/DFS/raft"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.InfoLevel)
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05.00000"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	rand.Seed(time.Now().UnixNano())
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
}
