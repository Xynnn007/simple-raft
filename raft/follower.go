package raft

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func (n *Node) startFollower() {
	if n.State == FOLLOWER {
		return
	}
	n.VoteFor = -1
	if n.State == CANDIDATE {
		n.exitCandidate <- 1
	}

	if n.State == LEADER {
		n.exitLeader <- 1
	}

	n.State = FOLLOWER
	log.Infof("Start follower..")
	n.Timer.Reset(CHECKLEADER_INTERVAL * time.Millisecond)

	<-n.Timer.C
	n.CurrentTerm++
	log.Infof("Timeout for heartbeat, turn into Candidate with term %v", n.CurrentTerm)
	n.startCandidate()
}

func (n *Node) refreshHeartbeat() {
	log.Infof("Get heartbeat from leader %v", n.Leader)
	n.Timer.Reset(CHECKLEADER_INTERVAL * time.Millisecond)
}
