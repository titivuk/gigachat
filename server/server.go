package server

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"unicode/utf8"

	"github.com/titivuk/gigachat/v2/common"
)

const (
	SERVER_SENDER    = "[Server]"
	AUTHORISED_MSG   = "Authorised"
	UNAUTHORISED_MSG = "Unauthorised"
	DISCONNECTED_MSG = "Disconnected"
)

func NewServer(token string) Server {
	return Server{
		Token:   token,
		clients: map[string]Client{},
	}
}

type Server struct {
	Token     string
	clients   map[string]Client
	clientsMu sync.Mutex
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	client := NewClient(conn)

	for {
		var msg common.Message
		err := client.Decoder.Decode(&msg)

		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client disconnected: %s\n", client.sender())

				s.RemoveClient(client)
				disconnectMsg := common.Message{
					Type:    common.MSG_TYPE,
					Payload: DISCONNECTED_MSG,
					Sender:  client.sender(),
				}
				s.BroadcastMessage(disconnectMsg, client)
				return
			}

			log.Printf("Cannot read message from %s. Error: %s", client.sender(), err)
			continue
		}

		log.Printf("Incoming message: %v\n", msg)

		switch msg.Type {
		case common.AUTH_TYPE:
			if client.Authorised {
				continue
			}

			if msg.Payload != s.Token {
				unauthMsg := common.Message{Type: common.AUTH_TYPE, Payload: UNAUTHORISED_MSG, Sender: SERVER_SENDER}
				s.SendMessage(client, unauthMsg)
				s.RemoveClient(client)

				disconnectMsg := common.Message{
					Type:    common.MSG_TYPE,
					Payload: DISCONNECTED_MSG,
					Sender:  client.sender(),
				}
				s.BroadcastMessage(disconnectMsg, client)
				return
			}

			client.Authorised = true
			client.username = msg.Sender
			s.AddClient(client)

			joinMsg := common.Message{
				Type:    common.MSG_TYPE,
				Payload: "Joined",
				Sender:  client.sender(),
			}
			s.BroadcastMessage(joinMsg, client)

			authMsg := common.Message{Type: common.AUTH_TYPE, Payload: AUTHORISED_MSG, Sender: SERVER_SENDER}
			s.SendMessage(client, authMsg)
		case common.MSG_TYPE:
			if !client.Authorised {
				responseMsg := common.Message{Type: common.AUTH_TYPE, Payload: UNAUTHORISED_MSG, Sender: SERVER_SENDER}
				s.SendMessage(client, responseMsg)
				s.RemoveClient(client)
				return
			}

			if msg.Payload == "" || !utf8.ValidString(msg.Payload) {
				responseMsg := common.Message{Type: common.ERROR_TYPE, Payload: "msg.Payload should be a valid non-empty utf-8 string", Sender: SERVER_SENDER}
				s.SendMessage(client, responseMsg)
				continue
			}

			msg.Sender = client.sender()
			s.BroadcastMessage(msg, client)
		default:
			fmt.Printf("Unknown message type: %d\n", msg.Type)
		}
	}
}

func (s *Server) AddClient(client Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients[client.sender()] = client
}

func (s *Server) RemoveClient(client Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	delete(s.clients, client.sender())
}

func (s *Server) BroadcastMessage(msg common.Message, source Client) {
	for _, c := range s.clients {
		if c != source {
			s.SendMessage(c, msg)
		}
	}
}

func (s *Server) SendMessage(target Client, msg common.Message) {
	err := target.Encoder.Encode(msg)
	if err != nil {
		fmt.Printf("Could not send message to connection: %s", err)
	}
}

func NewClient(conn net.Conn) Client {
	return Client{
		Conn:    conn,
		Encoder: gob.NewEncoder(conn),
		Decoder: gob.NewDecoder(conn),
	}
}

type Client struct {
	Authorised bool
	Conn       net.Conn
	Encoder    *gob.Encoder
	Decoder    *gob.Decoder
	username   string
}

func (c *Client) sender() string {
	return c.username
}
