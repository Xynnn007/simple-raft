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
	Term int
	*VoteRequest
	*VoteResponse
	*AppendRequest
	*AppendResponse
	*ClientRequest
}

type VoteRequest struct {
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type VoteResponse struct {
	VoteGranted bool
}

type AppendRequest struct {
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Entry
	LeaderCommit int
}

type AppendResponse struct {
	Id      int
	Index   int
	Success bool
}

type ClientRequest struct {
	Content
}

func (n *Node) handleVoteReq(v *VoteRequest, term int) error {
	log.Debugf("Vote Request: %v", v)
	res := Message{
		Type: VOTE_RES,
		Term: n.CurrentTerm,
		VoteResponse: &VoteResponse{
			VoteGranted: false,
		},
	}

	if n.CurrentTerm > term {
		log.Debugf("Do not vote for candidate %v of term %v < %v", v.CandidateId, term, n.CurrentTerm)

		data, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("reject vote failed: %v", err)
		}

		log.Debugf("Send voteres data: %v", string(data))
		n.send(v.CandidateId, data, "Reject vote")

		return nil
	}

	if (n.VoteFor == -1 || n.VoteFor == v.CandidateId) && v.LastLogIndex >= n.LogIndex {
		log.Infof("Vote for candidate %v of term %v.", v.CandidateId, term)
		n.VoteFor = v.CandidateId

		go n.startFollower()
		res.Term = term
		res.VoteResponse.VoteGranted = true

		data, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("vote failed: %v", err)
		}

		log.Debugf("Send voteres data: %v", string(data))
		n.send(v.CandidateId, data, "Vote")

		return nil
	}

	data, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("reject vote failed: %v", err)
	}

	log.Debugf("Send voteres data: %v", string(data))
	n.send(v.CandidateId, data, "Reject vote")

	return nil
}

func (n *Node) handleVoteRes(v *VoteResponse, term int) {
	log.Debugf("Vote Response: %v", *v)
	if n.State != CANDIDATE {
		return
	}

	if term > n.CurrentTerm {
		go n.startFollower()
		return
	}

	if v.VoteGranted {
		log.Infof("Get vote")
		n.getVotes++
		if n.getVotes > n.TotalPeersCount/2 {
			go n.startLeader()
		}
	}
	log.Debug("But do nothing...")
}

func (n *Node) handleAppendReq(v *AppendRequest, term int) error {
	var err error
	res := Message{
		Type: APPEND_RES,
		Term: n.CurrentTerm,
		AppendResponse: &AppendResponse{
			Id:      n.Myself.Id,
			Success: true,
		},
	}

	if n.CurrentTerm > term {
		res.AppendResponse.Success = false
		data, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("send reject appendRequest failed: %v", err)
		}

		log.Debugf("Send appendres data: %v", string(data))
		n.send(v.LeaderId, data, "Send reject appendRequest")

		return nil
	}

	if n.State != FOLLOWER {
		n.Leader = v.LeaderId
		log.Infof("Start following leader %v.", v.LeaderId)
		go n.startFollower()
	} else {
		n.Leader = v.LeaderId
		n.refreshHeartbeat()
	}

	index := n.Records.Find(v.PrevLogTerm, v.PrevLogIndex)
	if index == -1 {
		res.AppendResponse.Success = false
		data, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("send reject appendRequest failed: %v", err)
		}

		log.Debugf("Send appendres data: %v", string(data))
		n.send(v.LeaderId, data, "Send reject appendRequest")

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

	res.AppendResponse.Index = n.LogIndex
	data, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("send accept appendRequest failed: %v", err)
	}

	log.Debugf("Send appendres data: %v", string(data))
	n.send(v.LeaderId, data, "Send accept appendRequest")

	return err
}

func (n *Node) handleAppendRes(v *AppendResponse, term int) {
	if v.Success {
		n.nextIndex[v.Id] = v.Index + 1
		n.matchIndex[v.Id] = v.Index
		n.updateCommit()
		return
	}

	if term > n.CurrentTerm {
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
