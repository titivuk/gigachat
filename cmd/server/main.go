package main

import (
	"fmt"
	"log"
	"net"

	"github.com/titivuk/gigachat/v2/server"
)

const (
	PORT = 8080
)

func main() {
	log.Println("Starting server...")

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatal("Could not start server")
	}
	defer ln.Close()
	log.Printf("Server is listening on port %d...", PORT)

	server := server.NewServer(
		// server.GenServerToken()
		"abc",
	)

	log.Printf("Server token: %s", server.Token)

	for {
		conn, err := ln.Accept()
		log.Println("Incoming connection")

		if err != nil {
			log.Printf("Could not accept connection: %s", err)
			continue
		}

		go server.HandleConnection(conn)
	}
}
