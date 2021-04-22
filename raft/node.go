package raft

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Xynnn007/DFS/config"
)

const (
	HEARTBEAT_INTERVAL   = 1
	CHECKLEADER_INTERVAL = 5
)

type Transporter interface {
	Send(address string, port string, data []byte) error
	Recv() chan []byte
}

type Node struct {
	Peers           map[int]*config.Peer
	Myself          *config.Peer
	Leader          int
	TotalPeersCount int

	VoteFor     int
	CurrentTerm int
	State       int
	LogIndex    int
	Records     Entries

	commitIndex int
	lastApplied int
	getVotes    int

	nextIndex  map[int]int
	matchIndex map[int]int

	exitLeader    chan int
	exitCandidate chan int

	TransInterface Transporter
	Timer          *time.Ticker
}

func New(nc *config.NodeConfig) *Node {
	n := &Node{
		Peers: make(map[int]*config.Peer),
		Myself: &config.Peer{
			Address: nc.Address,
			Port:    nc.Port,
			Id:      nc.Id,
		},
		Leader: -1,

		VoteFor:     -1,
		CurrentTerm: 0,

		State:    FOLLOWER,
		LogIndex: 0,
		Records:  Entries{},

		commitIndex: 0,
		lastApplied: 0,
		getVotes:    0,

		nextIndex:  make(map[int]int),
		matchIndex: make(map[int]int),

		exitLeader:    make(chan int),
		exitCandidate: make(chan int),

		Timer: time.NewTicker(CHECKLEADER_INTERVAL * time.Second),
	}
	for _, v := range nc.Peers {
		n.Peers[v.Id] = &config.Peer{
			Address: v.Address,
			Port:    v.Port,
			Id:      v.Id,
		}
	}

	for k := range n.Peers {
		log.Debugf("New peer, id: %v --> %v", k, *n.Peers[k])
	}
	n.TotalPeersCount = len(n.Peers)

	return n
}

func (n *Node) SetTransporter(t Transporter) {
	n.TransInterface = t
}

func (n *Node) Start() {
	go n.startFollower()
	recv := n.TransInterface.Recv()

	for {
		message, ok := <-recv
		if !ok {
			log.Error("Failed to get message from channel.")
			continue
		}

		body := &Message{}
		err := json.Unmarshal(message, body)
		if err != nil {
			log.Errorf("Failed to get message : %v.", err)
			continue
		}
		log.Debugf("Handle message : %v", *body)
		n.handle(body)
	}
}

func (n *Node) handle(m *Message) {
	switch m.Type {
	case VOTE_REQ:
		n.handleVoteReq(m.VoteRequest)
	case VOTE_RES:
		n.handleVoteRes(m.VoteResponse)
	case APPEND_REQ:
		n.handleAppendReq(m.AppendRequest)
	case APPEND_RES:
		n.handleAppendRes(m.AppendResponse)
	default:
		n.handleClientReq(m.ClientRequest)
	}
}

func (n *Node) send(id int, data []byte) error {
	address := n.Peers[id].Address
	port := n.Peers[id].Port
	return n.TransInterface.Send(address, port, data)
}

func (n *Node) handleClientReq(r *ClientRequest) {
	if n.State != LEADER {
		js := Message{
			Type:          CLIENT_REQ,
			ClientRequest: r,
		}
		data, err := json.Marshal(js)
		if err != nil {
			log.Errorf("Forward client request to leader failed: %v", err)
			return
		}

		err = n.send(n.Leader, data)
		if err != nil {
			log.Errorf("Forward client request to leader %v failed: %v", n.Leader, err)
			return
		}
		log.Infof("Forward client request to leader %v", n.Leader)
	} else {
		n.LogIndex++
		e := &Entry{
			Term:    n.CurrentTerm,
			Index:   n.LogIndex,
			Content: r.Content,
		}
		n.Records.Insert(e)
		log.Infof("New client request Term: %v, Index: %v", n.CurrentTerm, n.LogIndex)
	}
}
