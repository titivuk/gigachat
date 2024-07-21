package client

func StartClient() {
	token := "abc"

	ui := NewUi()
	go ui.Start()

	conn := NewConnection(token)
	go conn.connect()

	for {
		select {
		case msg := <-ui.msg:
			// server does not send message back to sender
			ui.addMessage(msg)
			conn.sendMessage(msg)
		case msg := <-conn.msg:
			ui.addMessage(msg)
		}
	}
}
