package node

const (
	HEARTBEAT = 0
	MESSAGE   = 1
)

type message struct {
	ttype   int
	id      int
	content string
}
