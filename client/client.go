package client

func StartClient(token, username string) {
	ui := NewUi(username)
	go ui.Start()

	conn := NewConnection(token, username)
	go conn.connect()

	for {
		select {
		case msg := <-ui.msg:
			ui.addMessage(msg)
			conn.sendMessage(msg)
		case msg := <-conn.msg:
			ui.addMessage(msg)
		}
	}
}
