package main

import (
	"flag"

	"github.com/titivuk/gigachat/v2/client"
)

func main() {
	var token, username string
	flag.StringVar(&token, "t", "", "token to connect to server")
	flag.StringVar(&username, "u", "", "username to be displayed")
	flag.Parse()

	client.StartClient(token, username)
}
