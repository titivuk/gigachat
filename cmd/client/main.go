package main

import (
	"flag"

	"github.com/titivuk/gigachat/v2/client"
)

func main() {
	var token, username, serverUrl string
	flag.StringVar(&token, "t", "", "token to connect to server")
	flag.StringVar(&username, "u", "", "username to be displayed")
	flag.StringVar(&serverUrl, "s", ":8080", "server url")
	flag.Parse()

	client.StartClient(token, username, serverUrl)
}
