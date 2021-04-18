package raft

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

func (n *Node) startLeader() {
	if n.State == CANDIDATE {
		n.exitCandidate <- 1
	}

	n.State = LEADER

	n.Timer.Stop()
	// reset nextIndex
	for k, _ := range n.nextIndex {
		n.nextIndex[k] = n.LogIndex + 1
		n.matchIndex[k] = 0
	}

	// heartbeat timer
	n.Timer.Reset(HEARTBEAT_INTERVAL * time.Second)

	for {
		select {
		case <-n.exitLeader:
			return
		case <-n.Timer.C:
			go n.sendHeartbeat()
		}
	}

}

func (n *Node) updateCommit() {
	// update commitIndex
}

func (n *Node) stopLeader() {
	n.exitLeader <- 1
}

func (n *Node) sendHeartbeat() {
	for id := range n.Peers {
		go func(id int) {
			req := AppendRequest{
				Term:         n.CurrentTerm,
				LeaderId:     n.Myself.Id,
				LeaderCommit: n.commitIndex,
				PrevLogIndex: n.matchIndex[id],
			}

			term := n.Records.FindTerm(req.PrevLogIndex)

			if term == -1 {
				log.Errorf("Can not find entry of id %v to peer %v", req.PrevLogIndex, id)
				return
			}

			req.PrevLogTerm = term
			req.Entries = n.Records.GetRange(n.nextIndex[id], ENTRY_PER_REQ)

			data, err := json.Marshal(req)
			if err != nil {
				log.Errorf("Send heartbeat to %v failed: %v", id, err)
				return
			}

			err = n.send(id, data)
			if err != nil {
				log.Errorf("Send heartbeat to %v failed: %v", id, err)
			}
		}(id)
	}
}
