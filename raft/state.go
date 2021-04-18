package raft

const (
	LEADER    = 0
	CANDIDATE = 1
	FOLLOWER  = 2
)

var StateStr map[int]string

func init() {
	StateStr = make(map[int]string)
	StateStr[LEADER] = "Leader"
	StateStr[CANDIDATE] = "Candidate"
	StateStr[FOLLOWER] = "Follower"
}
