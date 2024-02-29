package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

const (
	_ = iota
	AUTH
	MSG
	ERROR
)

var rows = 0

const CLIENT_SENDER = "[Client]"

type Client struct {
	Conn    net.Conn
	Encoder *gob.Encoder
	Decoder *gob.Decoder
}

type Message struct {
	Type    int
	Sender  string
	Payload string
}

func main() {
	clearScreen()
	moveCursorTo(0, 0)

	printMessage(CLIENT_SENDER, fmt.Sprintf("window size: %v", getWinSize()))

	var token string
	flag.StringVar(&token, "t", "", "Token to connect to server")
	flag.Parse()

	if token == "" {
		printMessage(CLIENT_SENDER, "Missing token")
		return
	}

	printMessage(CLIENT_SENDER, "Connecting to the sever...")
	conn, connErr := net.Dial("tcp", ":8080")
	if connErr != nil {
		printMessage(CLIENT_SENDER, fmt.Sprintf("Could not connect to a server - %s", connErr))
		return
	}
	defer conn.Close()
	printMessage(CLIENT_SENDER, "Connected")
	printMessage(CLIENT_SENDER, fmt.Sprintf("My address - %s", conn.LocalAddr().String()))

	client := Client{
		Conn:    conn,
		Encoder: gob.NewEncoder(conn),
		Decoder: gob.NewDecoder(conn),
	}

	msg := Message{
		Type:    AUTH,
		Payload: token,
		Sender:  client.Conn.LocalAddr().String(),
	}
	sendAuthErr := client.Encoder.Encode(msg)
	if sendAuthErr != nil {
		printMessage(CLIENT_SENDER, fmt.Sprintf("Could not send message - %s", sendAuthErr))
		return
	}

	go handleClientInput(client)

	for {
		var msg Message
		err := client.Decoder.Decode(&msg)
		if err == io.EOF {
			printMessage(CLIENT_SENDER, "Connection closed")
			return
		}
		if err != nil {
			printMessage(CLIENT_SENDER, fmt.Sprintf("Error decoding message - %s", err))
			return
		}

		switch msg.Type {
		case AUTH:
			printMessage(msg.Sender, msg.Payload)

			if msg.Payload == "Unauthorised" {
				return
			}
		case MSG:
			printMessage(msg.Sender, msg.Payload)
		case ERROR:
			printMessage(msg.Sender, msg.Payload)
		}
	}
}

func handleClientInput(client Client) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		payload := scanner.Text()
		rows++

		if payload == "" || !utf8.ValidString(payload) {
			rows--
			fmt.Print("\033[1A")
			clearLine()
			continue
		}

		msg := Message{
			Type:    MSG,
			Payload: payload,
			Sender:  client.Conn.LocalAddr().String(),
		}
		err := client.Encoder.Encode(msg)
		if err != nil {
			printMessage(CLIENT_SENDER, fmt.Sprintf("Could not send message - %s", err))
		}
	}

	if err := scanner.Err(); err != nil {
		printMessage(CLIENT_SENDER, fmt.Sprintf("Error reading stdin - %s", err))
	}
}

func printMessage(sender string, msg string) {
	rows++
	fmt.Println(fmt.Sprintf("[%d]", rows), sender, ":", msg)
}

func clearLine() {
	fmt.Print("\033[K")
}

func clearScreen() {
	fmt.Print("\033[2J")
}

func moveCursorTo(x, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getWinSize() *winsize {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}

	return ws
}
