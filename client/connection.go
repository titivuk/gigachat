package client

import (
	"encoding/gob"
	"io"
	"log"
	"net"

	"github.com/titivuk/gigachat/v2/common"
)

const (
	_ = iota
	AUTH
	MSG
	ERROR
)

const CLIENT_SENDER = "[Client]"

func NewConnection(token string) Connection {

	return Connection{
		msg:   make(chan string),
		token: token,
	}
}

type Connection struct {
	conn    net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
	msg     chan string
	token   string
}

func (c *Connection) connect() {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to connect to the server")
	}
	defer conn.Close()

	c.conn = conn
	c.encoder = gob.NewEncoder(conn)
	c.decoder = gob.NewDecoder(conn)

	msg := common.Message{
		Type:    AUTH,
		Payload: c.token,
		Sender:  c.sender(),
	}
	err = c.encoder.Encode(msg)
	if err != nil {
		log.Fatalf("Failed to send auth message - %s", err)
	}

	for {
		var msg common.Message
		err := c.decoder.Decode(&msg)
		if err == io.EOF {
			c.msg <- msg.Payload
			log.Fatal("Connection closed")
		}
		if err != nil {
			c.msg <- msg.Payload
			log.Fatalf("Error decoding message - %s", err)
		}

		switch msg.Type {
		case AUTH:
			c.msg <- msg.Payload

			if msg.Payload == "Unauthorised" {
				log.Fatal(msg.Payload)
			}
		case MSG:
			c.msg <- msg.Payload
		case ERROR:
			c.msg <- msg.Payload
		}
	}
}

func (c *Connection) sendMessage(payload string) error {
	msg := common.Message{
		Type:    MSG,
		Payload: payload,
		Sender:  c.sender(),
	}

	return c.encoder.Encode(msg)
}

func (c *Connection) sender() string {
	return c.conn.LocalAddr().String()
}
