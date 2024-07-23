package client

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/gdamore/tcell"
	"github.com/titivuk/gigachat/v2/common"
)

const (
	MAX_MSG_LENGTH = 256
	MIN_WIDTH      = 50
	MIN_HEIGHT     = 6
)

func NewUi(username string) *Ui {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("+%v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("+%v", err)
	}

	width, height := screen.Size()

	mi := NewMessageInput(
		0, height-1-2,
		width-1, height-1,
		MAX_MSG_LENGTH,
	)
	chatHistory := NewChatHistory(
		0, 0,
		width-1, mi.Y1-1,
	)

	return &Ui{
		messageInput: *mi,
		chatHistory:  *chatHistory,
		msg:          make(chan common.Message),
		screen:       screen,
		username:     username,
	}
}

type Ui struct {
	messageInput MessageInput
	chatHistory  ChatHistory
	msg          chan common.Message
	screen       tcell.Screen
	username     string
}

func (ui *Ui) Start() {
	ui.screen.Clear()
	ui.screen.Show()

	for {
		ui.render()

		event := ui.screen.PollEvent()
		switch ev := event.(type) {
		case *tcell.EventResize:
			ui.screen.Sync()
		// render(ui.screen, chatHistory, mi)
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				ui.screen.Fini()
				os.Exit(0)
			} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				ui.messageInput.removeChar()
			} else if ev.Key() == tcell.KeyEnter {
				// send message
				if len(ui.messageInput.Text) > 0 {
					ui.msg <- common.Message{
						Payload: string(ui.messageInput.Text),
						Type:    common.MSG_TYPE,
						Sender:  ui.username,
					}
					ui.messageInput.clear()
				}
			} else if 32 <= ev.Rune() && ev.Rune() <= 126 {
				ui.messageInput.addChar(ev.Rune())
			}
		}
	}
}

func (ui *Ui) addMessage(msg common.Message) {
	ui.chatHistory.addMessage(msg)
	ui.render()
}

func (ui *Ui) recalculateSize() {
	width, height := ui.screen.Size()
	width = max(width, MIN_WIDTH)
	height = max(height, MIN_HEIGHT)

	ui.messageInput.X2 = width - 1
	ui.messageInput.Y2 = height - 1
	rowCapacity := ui.messageInput.rowCapacity()
	ui.messageInput.Y1 = ui.messageInput.Y2 - 1 - max(
		1, // when there is no text we still want empty row
		int(math.Ceil(float64(len(ui.messageInput.Text))/float64(rowCapacity))), // +1 for char counter row
	)

	ui.chatHistory.X2 = width - 1
	ui.chatHistory.Y2 = ui.messageInput.Y1 - 1
}

func (ui *Ui) render() {
	ui.screen.Clear()
	defer ui.screen.Show()

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	ui.recalculateSize()
	ui.renderChatHistory(style)
	ui.renderMessageInput(style)
}

func (ui *Ui) renderChatHistory(style tcell.Style) {
	drawBox(ui.screen, ui.chatHistory.X1, ui.chatHistory.Y1, ui.chatHistory.X2, ui.chatHistory.Y2, style)

	// calcute message idx from which we should render chat history
	rowCapacity := ui.chatHistory.rowCapacity()
	startY := ui.chatHistory.Y1 + 1
	startMsgIdx := 0
	var msg string
	var msgLen, rows int
	for i := len(ui.chatHistory.messages) - 1; i >= 0; i-- {
		msg = fmt.Sprintf("%s: %s", ui.chatHistory.messages[i].Sender, ui.chatHistory.messages[i].Payload)
		msgLen = len(msg)
		rows = int(math.Ceil(float64(msgLen) / float64(rowCapacity)))

		// not enough space. do not render message
		if startY+rows-1 >= ui.chatHistory.Y2 {
			break
		}

		startY += rows
		startMsgIdx = i
	}

	// render every message starting from startMsgIdx
	startY = ui.chatHistory.Y1 + 1
	for i := startMsgIdx; i < len(ui.chatHistory.messages); i++ {
		msg = fmt.Sprintf("%s: %s", ui.chatHistory.messages[i].Sender, ui.chatHistory.messages[i].Payload)
		msgLen = len(msg)

		rows = int(math.Ceil(float64(msgLen) / float64(rowCapacity)))
		drawText(ui.screen,
			ui.chatHistory.X1+1, startY,
			ui.chatHistory.X2-1, startY+rows-1,
			style,
			msg,
		)

		startY += rows
	}
}

func (ui *Ui) renderMessageInput(style tcell.Style) {
	// draw input message box
	drawBox(
		ui.screen,
		ui.messageInput.X1, ui.messageInput.Y1,
		ui.messageInput.X2, ui.messageInput.Y2,
		style,
	)
	// draw input message text
	drawText(
		ui.screen,
		ui.messageInput.X1+1, ui.messageInput.Y1+1,
		ui.messageInput.X2-1, ui.messageInput.Y2-1,
		style,
		string(ui.messageInput.Text),
	)
	// draw cursor
	ui.screen.ShowCursor(ui.messageInput.cursorPosition())
	// draw characters counter
	counter := fmt.Sprintf("%d/%d", len(ui.messageInput.Text), ui.messageInput.maxLen)
	drawText(
		ui.screen,
		ui.messageInput.X2-len(counter), ui.messageInput.Y2,
		ui.messageInput.X2, ui.messageInput.Y2,
		style,
		counter,
	)
}

func NewMessageInput(x1, y1, x2, y2 int, maxLen uint) *MessageInput {
	return &MessageInput{
		X1:     x1,
		Y1:     y1,
		X2:     x2,
		Y2:     y2,
		Text:   make([]rune, 0),
		maxLen: maxLen,
	}
}

type MessageInput struct {
	X1, Y1 int
	X2, Y2 int
	Text   []rune
	maxLen uint
}

func (mi *MessageInput) addChar(ch rune) bool {
	if len(mi.Text) < int(mi.maxLen) {
		mi.Text = append(mi.Text, ch)
		return true
	}

	return false
}

func (mi *MessageInput) removeChar() bool {
	if len(mi.Text) > 0 {
		mi.Text = mi.Text[:len(mi.Text)-1]
		return true
	}

	return false
}

func (mi *MessageInput) clear() {
	mi.Text = make([]rune, 0)
}

func (mi *MessageInput) rowCapacity() int {
	return mi.X2 - mi.X1 - 1
}

func (mi *MessageInput) cursorPosition() (int, int) {
	rowCapacity := mi.rowCapacity()

	lastRowLength := len(mi.Text) % rowCapacity
	// if last string length equals to rowCapacity we want cursor to be on the same row
	if len(mi.Text) > 0 && lastRowLength == 0 {
		lastRowLength = len(mi.Text)
	}

	return mi.X1 + lastRowLength + 1, mi.Y2 - 1
}

func NewChatHistory(x1, y1, x2, y2 int) *ChatHistory {
	return &ChatHistory{
		X1:       x1,
		Y1:       y1,
		X2:       x2,
		Y2:       y2,
		messages: make([]common.Message, 0),
	}
}

type ChatHistory struct {
	X1, Y1   int
	X2, Y2   int
	messages []common.Message
}

func (ch *ChatHistory) rowCapacity() int {
	return ch.X2 - ch.X1 - 1
}

func (ch *ChatHistory) addMessage(msg common.Message) {
	ch.messages = append(ch.messages, msg)
}

func drawBox(screen tcell.Screen, x1, y1, x2, y2 int, style tcell.Style) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	for dx := x1 + 1; dx < x2; dx++ {
		screen.SetContent(dx, y1, tcell.RuneHLine, nil, style)
		screen.SetContent(dx, y2, tcell.RuneHLine, nil, style)
	}

	for dy := y1 + 1; dy < y2; dy++ {
		screen.SetContent(x1, dy, tcell.RuneVLine, nil, style)
		screen.SetContent(x2, dy, tcell.RuneVLine, nil, style)
	}

	if x1 != x2 || y1 != y2 {
		screen.SetContent(x1, y1, tcell.RuneULCorner, nil, style)
		screen.SetContent(x2, y1, tcell.RuneURCorner, nil, style)
		screen.SetContent(x2, y2, tcell.RuneLRCorner, nil, style)
		screen.SetContent(x1, y2, tcell.RuneLLCorner, nil, style)
	}
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range text {
		s.SetContent(col, row, r, nil, style)
		col++
		if col > x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}
