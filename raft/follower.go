package raft

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func (n *Node) startFollower() {
	log.Infof("Start follower..")
	if n.State == CANDIDATE {
		n.exitCandidate <- 1
	}

	if n.State == LEADER {
		n.exitLeader <- 1
	}

	n.State = FOLLOWER

	n.Timer.Reset(CHECKLEADER_INTERVAL * time.Second)

	<-n.Timer.C
	log.Info("Timeout for heartbeat, turn into Candidate")
	n.startCandidate()
}

func (n *Node) refreshHeartbeat() {
	log.Infof("Get heartbeat from leader %v", n.Leader)
	n.Timer.Reset(CHECKLEADER_INTERVAL * time.Second)
}
