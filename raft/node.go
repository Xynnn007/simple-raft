package raft

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Xynnn007/simple-raft/config"
)

const (
	HEARTBEAT_INTERVAL   = 300  //ms
	CHECKLEADER_INTERVAL = 1000 //ms
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

		State:    INIT,
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
		log.Debugf("Handle message : %v", string(message))
		n.handle(body)
	}
}

func (n *Node) handle(m *Message) {
	switch m.Type {
	case VOTE_REQ:
		n.handleVoteReq(m.VoteRequest, m.Term)
	case VOTE_RES:
		n.handleVoteRes(m.VoteResponse, m.Term)
	case APPEND_REQ:
		n.handleAppendReq(m.AppendRequest, m.Term)
	case APPEND_RES:
		n.handleAppendRes(m.AppendResponse, m.Term)
	default:
		n.handleClientReq(m.ClientRequest)
	}
}

func (n *Node) send(id int, data []byte, errmsg string) {
	address := n.Peers[id].Address
	port := n.Peers[id].Port
	go func() {
		err := n.TransInterface.Send(address, port, data)
		if err != nil {
			log.Errorf("Send to %v failed, content : %v. error:%v:%v", id, string(data), errmsg, err)
		}
	}()
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

		n.send(n.Leader, data, "Forward client request to leader")
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
