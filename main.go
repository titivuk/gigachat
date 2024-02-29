package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sync"
	"unicode/utf8"
)

const (
	PORT             = 8080
	SERVER_SENDER    = "[Server]"
	AUTHORISED_MSG   = "Authorised"
	UNAUTHORISED_MSG = "Unauthorised"
)

const (
	_ = iota
	AUTH
	MSG
	ERROR
)

var (
	base58Alphabet = []rune{'1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
)

func GenServerToken() string {
	randomInt := rand.Uint64()

	var buf bytes.Buffer
	var remainder uint64
	alphabetLen := uint64(len(base58Alphabet))
	for randomInt != 0 {
		remainder = randomInt % alphabetLen
		buf.WriteRune(base58Alphabet[remainder])
		randomInt /= alphabetLen
	}

	return buf.String()
}

type Message struct {
	Type    int
	Payload string
	Sender  string
}

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

func (s *Server) AddClient(client Client) {
	s.clientsMu.Lock()
	s.clients[client.Conn.RemoteAddr().String()] = client
	s.clientsMu.Unlock()
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	client := NewClient(conn)

	for {
		var msg Message
		err := client.Decoder.Decode(&msg)
		if err == io.EOF {
			fmt.Printf("Client disconnected: %s\n", client.Conn.RemoteAddr().String())
			s.RemoveClient(client)
			disconnectMsg := Message{
				Type:    MSG,
				Payload: "Disconnected",
				Sender:  client.Conn.RemoteAddr().String(),
			}
			s.BroadcastMessage(disconnectMsg, client)
			return
		}
		if err != nil {
			fmt.Printf("Could not read message from %s: %s\n", client.Conn.RemoteAddr().String(), err)
			s.RemoveClient(client)
			disconnectMsg := Message{
				Type:    MSG,
				Payload: "Disconnected",
				Sender:  client.Conn.RemoteAddr().String(),
			}
			s.BroadcastMessage(disconnectMsg, client)
			return
		}
		fmt.Printf("Incoming message: %v\n", msg)

		switch msg.Type {
		case AUTH:
			if client.Authorised {
				continue
			}

			if msg.Payload != s.Token {
				unauthMsg := Message{Type: AUTH, Payload: UNAUTHORISED_MSG, Sender: SERVER_SENDER}
				s.SendMessage(client, unauthMsg)
				s.RemoveClient(client)

				disconnectMsg := Message{
					Type:    MSG,
					Payload: "Disconnected",
					Sender:  client.Conn.RemoteAddr().String(),
				}
				s.BroadcastMessage(disconnectMsg, client)
				return
			}

			client.Authorised = true
			s.AddClient(client)

			joinMsg := Message{
				Type:    MSG,
				Payload: "Joined",
				Sender:  client.Conn.RemoteAddr().String(),
			}
			s.BroadcastMessage(joinMsg, client)

			authMsg := Message{Type: AUTH, Payload: AUTHORISED_MSG, Sender: SERVER_SENDER}
			s.SendMessage(client, authMsg)
		case MSG:
			if !client.Authorised {
				responseMsg := Message{Type: AUTH, Payload: UNAUTHORISED_MSG, Sender: SERVER_SENDER}
				s.SendMessage(client, responseMsg)
				s.RemoveClient(client)
				return
			}

			if msg.Payload == "" || !utf8.ValidString(msg.Payload) {
				responseMsg := Message{Type: ERROR, Payload: "msg.Payload should be a valid non-empty utf-8 string", Sender: SERVER_SENDER}
				s.SendMessage(client, responseMsg)
				continue
			}

			msg.Sender = client.Conn.RemoteAddr().String()
			s.BroadcastMessage(msg, client)
		default:
			fmt.Printf("Unknown message type: %d\n", msg.Type)
		}
	}
}

func (s *Server) RemoveClient(client Client) {
	s.clientsMu.Lock()
	delete(s.clients, client.Conn.RemoteAddr().String())
	s.clientsMu.Unlock()
}

func (s *Server) BroadcastMessage(msg Message, source Client) {
	for _, c := range s.clients {
		if c != source {
			s.SendMessage(c, msg)
		}
	}
}

func (s *Server) SendMessage(target Client, msg Message) {
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
}

func main() {
	fmt.Println("Starting server...")

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		panic("Could not start server")
	}
	defer ln.Close()
	fmt.Printf("Server is listening on port %d...\n", PORT)

	server := NewServer(GenServerToken())

	fmt.Fprintln(os.Stdout, server.Token)

	for {
		conn, err := ln.Accept()
		fmt.Println("Incoming connection")

		if err != nil {
			fmt.Printf("Could not accept connection: %s", err)
			continue
		}

		go server.HandleConnection(conn)
	}
}
