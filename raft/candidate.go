package raft

import (
	"encoding/json"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

func (n *Node) startCandidate() {
	n.State = CANDIDATE
	n.CurrentTerm++
	n.getVotes = 1
	n.VoteFor = n.Myself.Id

	log.Debug("Sending votereq...")
	n.sendVotereq()

	randTime := rand.Int63n(200) + 100
	log.Infof("Wait for %v ms as overtime", randTime)
	n.Timer.Reset(time.Duration(randTime) * time.Millisecond)

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
				Term: n.CurrentTerm,
				VoteRequest: &VoteRequest{

					CandidateId:  n.Myself.Id,
					LastLogIndex: n.LogIndex,
					LastLogTerm:  n.Records.FindTerm(n.LogIndex),
				},
			}
			req.Term = n.CurrentTerm
			if req.LastLogIndex == 0 {
				req.LastLogTerm = 0
			}
			data, err := json.Marshal(req)
			if err != nil {
				log.Errorf("Send vote request to %v failed: %v", index, err)
				return
			}

			log.Debugf("Send votereq to %v", index)
			n.send(index, data, "Send vote")
		}(id)
	}
}
