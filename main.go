package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gorilla/websocket"
)

type ChatMessage struct {
	Type      string    `json:"type"`
	Nick      string    `json:"nick"`
	To        string    `json:"to"`
	Msg       string    `json:"msg"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	TypeMsg           = "msg"
	TypeAddContact    = "add_contact"
	TypeRemoveContact = "remove_contact"
	TypeGetContact    = "get_contact"
	TypeContactList   = "contact_list"
)

var (
	nick string

	messagesContainer *fyne.Container
	scrollContainer   *container.Scroll
	friendListVbox    *fyne.Container
	targetEntry       *widget.Entry
	inputEntry        *widget.Entry
	addFriendEntry    *widget.Entry

	friendButtons = make(map[string]*widget.Button)
	mu            sync.Mutex
	conn          *websocket.Conn
)

func main() {

	fmt.Print("Send username: ")
	fmt.Scanln(&nick)

	var err error

	conn, _, err = websocket.DefaultDialer.Dial("ws://localhost:8443/ws", nil)
	if err != nil {
		fmt.Println("net.Dial err:", err)
		os.Exit(1)
	}
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(nick))
	if err != nil {
		fmt.Println("Error Nick:", err)
		return
	}

	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())

	myWindow := myApp.NewWindow("Test")
	myWindow.Resize(fyne.NewSize(400, 600))
	myWindow.CenterOnScreen()

	messagesContainer = container.NewVBox()
	scrollContainer = container.NewScroll(messagesContainer)
	scrollContainer.Resize(fyne.NewSize(400, 400))

	inputEntry = widget.NewEntry()
	inputEntry.SetPlaceHolder("Type a message...:")

	targetEntry = widget.NewEntry()
	targetEntry.SetPlaceHolder("To (Public if empty)")
	targetEntry.Disabled()

	sendBtn := widget.NewButton("Send", func() {
		text := inputEntry.Text
		if text == "" {
			return
		}

		msg := ChatMessage{
			Type: "msg",
			Msg:  text,
			To:   targetEntry.Text,
		}

		if err := conn.WriteJSON(msg); err != nil {
			fmt.Println("conn.WriteMessage err:", err)
		}
		inputEntry.SetText("")
	})

	inputContainer := container.NewBorder(nil, nil, nil, sendBtn, targetEntry, inputEntry)
	chatColumn := container.NewVBox(scrollContainer, inputContainer)

	addFriendEntry = widget.NewEntry()
	addFriendEntry.SetPlaceHolder("Friend Nick")
	addFriendBtn := widget.NewButton("Add", func() {
		newFriend := addFriendEntry.Text
		if newFriend == "" {
			return
		}

		msg := ChatMessage{
			Type: TypeAddContact,
			Msg:  newFriend,
		}
		if err := conn.WriteJSON(msg); err != nil {
			fmt.Println("conn.WriteMessage err:", err)
		}
		addFriendEntry.SetText("")
	})

	addFriendContainer := container.NewBorder(nil, nil, nil, addFriendBtn, addFriendEntry)

	friendListVbox = container.NewVBox()
	friendScroll := container.NewScroll(friendListVbox)
	friendScroll.SetMinSize(fyne.NewSize(200, 200))

	contactsColumn := container.NewVBox(
		widget.NewLabel("Contacts:"),
		addFriendContainer,
		widget.NewSeparator(),
		friendScroll,
	)

	content := container.NewGridWithColumns(2, chatColumn, contactsColumn)
	myWindow.SetContent(content)

	go listenForMessages(conn)

	requestContactList()

	myWindow.ShowAndRun()
}

func requestContactList() {
	msg := ChatMessage{Type: TypeGetContact, Nick: nick}
	if err := conn.WriteJSON(msg); err != nil {
		fmt.Println("Get contact err:", err)
	}
}

func listenForMessages(c *websocket.Conn) {
	for {
		var msg ChatMessage
		if err := c.ReadJSON(&msg); err != nil {
			fmt.Println("conn.ReadJSON err:", err)
			return
		}
		msgType := msg.Type
		if msgType == "" {
			msgType = "msg"
		}

		fmt.Printf("Received: Type=%s, From=%s, Payload=%s\n", msg.Type, msg.Nick, msg.Msg)

		switch msgType {
		case TypeMsg:
			addLabel(msg)
		case TypeContactList:
			var friends []string
			if err := json.Unmarshal([]byte(msg.Msg), &friends); err != nil {
				fmt.Println("Failed to parse list:", err)
				return
			}
			updateContactListUI(friends)
		case TypeAddContact:
			friendName := msg.To
			if friendName != "" {
				addSingleFriendToUI(friendName)
			}

		}
	}
}

func updateContactListUI(friends []string) {
	mu.Lock()
	defer mu.Unlock()

	friendListVbox.Objects = nil

	for _, name := range friends {
		createFriendButton(name)
	}
	friendListVbox.Refresh()
}

func addSingleFriendToUI(name string) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := friendButtons[name]; exists {
		return
	}
	createFriendButton(name)
	friendListVbox.Refresh()
}

func createFriendButton(name string) {
	btn := widget.NewButton(name, func() {
		if name == "General / Public" || name == "All" {
			targetEntry.SetText("")
		} else {
			targetEntry.SetText(name)
		}
		inputEntry.FocusLost()
	})
	friendButtons[name] = btn
	friendListVbox.Add(btn)
}

func addLabel(msg ChatMessage) {
	fyne.Do(func() {
		isPrivate := msg.To != ""
		isMine := msg.Nick == nick

		var headerText string
		var bubbleColor color.NRGBA

		if isPrivate {
			bubbleColor = color.NRGBA{R: 128, G: 0, B: 128, A: 255}

			if isMine {
				headerText = fmt.Sprintf("🔒 You ➔ %s", msg.To)
			} else {
				headerText = fmt.Sprintf("🔒 %s (Private)", msg.Nick)
			}
		} else {
			if isMine {
				bubbleColor = color.NRGBA{R: 30, G: 60, B: 150, A: 255}
				headerText = "You (Public)"
			} else {
				bubbleColor = color.NRGBA{R: 60, G: 60, B: 60, A: 255}
				headerText = msg.Nick
			}
		}

		headerLabel := widget.NewLabel(headerText)
		headerLabel.TextStyle.Bold = true
		headerLabel.Alignment = fyne.TextAlignLeading

		msgLabel := widget.NewLabel(msg.Msg)

		timeLabel := widget.NewLabel(msg.Timestamp.Format("15:04"))
		timeLabel.Alignment = fyne.TextAlignTrailing
		timeLabel.TextStyle.Italic = true

		bubbleContent := container.NewVBox(
			headerLabel,
			msgLabel,
			container.NewHBox(layout.NewSpacer(), timeLabel),
		)

		background := canvas.NewRectangle(bubbleColor)
		background.CornerRadius = 10

		bubble := container.NewStack(background, container.NewPadded(bubbleContent))

		var row *fyne.Container
		if isMine {
			row = container.NewHBox(layout.NewSpacer(), bubble)
		} else {
			row = container.NewHBox(bubble, layout.NewSpacer())
		}

		messagesContainer.Add(row)
		scrollContainer.ScrollToBottom()
		messagesContainer.Refresh()
	})
}
