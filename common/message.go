package common

type Message struct {
	Type    int
	Payload string
	Sender  string
}

const (
	_ = iota
	AUTH_TYPE
	MSG_TYPE
	ERROR_TYPE
)