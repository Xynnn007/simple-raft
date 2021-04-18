package raft

import (
	"encoding/json"
	"fmt"

	"github.com/Xynnn007/DFS/misc"
	log "github.com/sirupsen/logrus"
)

const (
	VOTE_REQ   = 0
	VOTE_RES   = 1
	APPEND_REQ = 2
	APPEND_RES = 3
	CLIENT_REQ = 4

	ENTRY_PER_REQ = 5
)

type Message struct {
	Type int
	*VoteRequest
	*VoteResponse
	*AppendRequest
	*AppendReponse
	*ClientRequest
}

type VoteRequest struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type VoteResponse struct {
	Term        int
	VoteGranted bool
}

type AppendRequest struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Entry
	LeaderCommit int
}

type AppendReponse struct {
	Term    int
	Id      int
	Index   int
	Success bool
}

type ClientRequest struct {
	Content
}

func (n *Node) handleVoteReq(v *VoteRequest) error {
	res := VoteResponse{
		Term:        n.CurrentTerm,
		VoteGranted: false,
	}

	if n.CurrentTerm > v.Term {
		log.Debugf("Do not vote for candidate %v of term %v < %v", v.CandidateId, v.Term, n.CurrentTerm)

		data, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("reject vote failed: %v", err)
		}

		err = n.send(v.CandidateId, data)
		if err != nil {
			return fmt.Errorf("reject vote failed: %v", err)
		}

		return nil
	}

	if (n.VoteFor == -1 || n.VoteFor == v.CandidateId) && v.LastLogIndex >= n.LogIndex {
		log.Infof("Vote for candidate %v of term %v.", v.CandidateId, v.Term)
		n.VoteFor = v.CandidateId

		n.startFollower()
		res.Term = v.Term
		res.VoteGranted = true

		data, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("vote failed: %v", err)
		}

		err = n.send(v.CandidateId, data)
		if err != nil {
			return fmt.Errorf("vote failed: %v", err)
		}

		return nil
	}

	data, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("reject vote failed: %v", err)
	}

	err = n.send(v.CandidateId, data)
	if err != nil {
		return fmt.Errorf("reject vote failed: %v", err)
	}

	return nil
}

func (n *Node) handleVoteRes(v *VoteResponse) {
	if n.State != CANDIDATE {
		return
	}

	if v.Term > n.CurrentTerm {
		n.startFollower()
		return
	}

	if v.VoteGranted {
		n.getVotes++
		if n.getVotes > n.TotalPeersCount/2 {
			n.startLeader()
		}
	}
}

func (n *Node) handleAppendReq(v *AppendRequest) error {
	var err error
	res := AppendReponse{
		Id:      n.Myself.Id,
		Term:    n.CurrentTerm,
		Success: true,
	}

	if n.CurrentTerm > v.Term {
		res.Success = false
		data, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("send reject appendRequest failed: %v", err)
		}

		err = n.send(v.LeaderId, data)
		if err != nil {
			return fmt.Errorf("send reject appendRequest failed: %v", err)
		}
		return nil
	}

	if n.State != FOLLOWER {
		n.Leader = v.LeaderId
		log.Infof("Start following leader %v.", v.LeaderId)
		go n.startFollower()
	} else {
		log.Infof("Get heartbeat from leader %v.", v.LeaderId)
		n.refreshHeartbeat()
	}

	index := n.Records.Find(v.PrevLogTerm, v.PrevLogIndex)
	if index == -1 {
		res.Success = false
		data, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("send reject appendRequest failed: %v", err)
		}

		err = n.send(v.LeaderId, data)
		if err != nil {
			return fmt.Errorf("send reject appendRequest failed: %v", err)
		}
		return nil
	}

	i := 0
	for ; i < len(v.Entries); i++ {
		if index < len(n.Records.Storage) && n.Records.Storage[index+1].Term == v.Entries[i].Term {
			index++
			continue
		} else {
			break
		}
	}

	if i < len(v.Entries) {
		log.Warnf("Entry at index %v conflicted, will delete log %v ~ %v.", v.Entries[i].Index,
			v.Entries[i].Index+1, v.Entries[len(v.Entries)-1].Index)

		n.Records.DeleteFrom(index)
		n.LogIndex = index
	}

	for i < len(v.Entries) {
		n.Records.Insert(&v.Entries[i])
		n.LogIndex++
		i++
	}

	if v.LeaderCommit > n.commitIndex {
		n.commitIndex = misc.Min(v.LeaderCommit, v.Entries[len(v.Entries)-1].Index)
	}

	res.Index = n.LogIndex
	data, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("send accept appendRequest failed: %v", err)
	}

	err = n.send(v.LeaderId, data)
	if err != nil {
		return fmt.Errorf("send accept appendRequest failed: %v", err)
	}
	return err
}

func (n *Node) handleAppendRes(v *AppendReponse) {
	if v.Success {
		n.nextIndex[v.Id] = v.Index + 1
		n.matchIndex[v.Id] = v.Index
		n.updateCommit()
		return
	}

	if v.Term > n.CurrentTerm {
		n.State = FOLLOWER
		n.stopLeader()
		return
	}

	if n.nextIndex[v.Id]-ENTRY_PER_REQ > n.commitIndex {
		n.nextIndex[v.Id] = n.nextIndex[v.Id] - ENTRY_PER_REQ
	} else {
		n.nextIndex[v.Id] = n.commitIndex
	}
}