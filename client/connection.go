package client

import (
	"encoding/gob"
	"io"
	"log"
	"net"

	"github.com/titivuk/gigachat/v2/common"
)

const CLIENT_SENDER = "[Client]"

func NewConnection(token, username string) Connection {
	return Connection{
		msg:      make(chan common.Message),
		token:    token,
		username: username,
	}
}

type Connection struct {
	conn     net.Conn
	encoder  *gob.Encoder
	decoder  *gob.Decoder
	msg      chan common.Message
	token    string
	username string
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
		Type:    common.AUTH_TYPE,
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
			c.msg <- msg
			log.Fatal("Connection closed")
		}
		if err != nil {
			c.msg <- msg
			log.Fatalf("Error decoding message - %s", err)
		}

		switch msg.Type {
		case common.AUTH_TYPE:
			c.msg <- msg

			if msg.Payload == "Unauthorised" {
				log.Fatal(msg.Payload)
			}
		case common.MSG_TYPE:
			c.msg <- msg
		case common.ERROR_TYPE:
			c.msg <- msg
		}
	}
}

func (c *Connection) sendMessage(msg common.Message) error {
	return c.encoder.Encode(msg)
}

func (c *Connection) sender() string {
	return c.username
}
