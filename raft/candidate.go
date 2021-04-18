package raft

import (
	"encoding/json"
	"math"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

func (n *Node) startCandidate() {
	n.State = CANDIDATE
	n.CurrentTerm++
	n.getVotes = 1
	n.VoteFor = n.Myself.Id

	log.Debug("Start send votereq...")
	n.sendVotereq()
	log.Debug("Send votereq done.")

	rand.Seed(time.Now().Unix())
	randTime := int64(1000.0*HEARTBEAT_INTERVAL*(rand.Int()/math.MaxInt32)) * 1000
	n.Timer.Reset(time.Duration(randTime))

	select {
	case <-n.exitCandidate:
		return
	case <-n.Timer.C:
		log.Warnf("Election timeout, start a new election")
		go n.startCandidate()
	}
}

func (n *Node) sendVotereq() {
	for id := range n.Peers {
		go func(id int) {
			req := VoteRequest{
				Term:         n.CurrentTerm,
				CandidateId:  n.Myself.Id,
				LastLogIndex: n.LogIndex,
				LastLogTerm:  n.Records.FindTerm(n.LogIndex),
			}

			data, err := json.Marshal(req)
			if err != nil {
				log.Errorf("Send vote request to %v failed: %v", id, err)
				return
			}

			log.Debugf("Send votereq to %v", id)
			err = n.send(id, data)
			if err != nil {
				log.Errorf("Send vote request to %v failed: %v", id, err)
			}
		}(id)
	}
}
