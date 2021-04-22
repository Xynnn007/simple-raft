package raft

import (
	"encoding/json"
	"math"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

func (n *Node) startCandidate() {
	if n.State == CANDIDATE {
		return
	}

	n.State = CANDIDATE
	n.CurrentTerm++
	n.getVotes = 1
	n.VoteFor = n.Myself.Id

	log.Debug("Sending votereq...")
	n.sendVotereq()

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
		go func(index int) {
			req := Message{
				Type: VOTE_REQ,
				VoteRequest: &VoteRequest{
					Term:         n.CurrentTerm,
					CandidateId:  n.Myself.Id,
					LastLogIndex: n.LogIndex,
					LastLogTerm:  n.Records.FindTerm(n.LogIndex),
				},
			}

			data, err := json.Marshal(req)
			if err != nil {
				log.Errorf("Send vote request to %v failed: %v", index, err)
				return
			}

			log.Debugf("Send votereq to %v", index)
			err = n.send(index, data)
			if err != nil {
				log.Errorf("Send vote request to %v failed: %v", index, err)
			}
		}(id)
	}
}
