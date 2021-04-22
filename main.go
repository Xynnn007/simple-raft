package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Xynnn007/DFS/config"
	"github.com/Xynnn007/DFS/network/http"
	"github.com/Xynnn007/DFS/raft"

	log "github.com/sirupsen/logrus"
)

var (
	BuildTime = ""
	CommitID  = ""
	GoVersion = ""
)

func init() {
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "-version") {
		fmt.Println("go version: \t" + GoVersion)
		fmt.Println("Build Time: \t" + BuildTime)
		fmt.Println("Program Commit ID : \t" + CommitID)
		os.Exit(1)
	}

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
