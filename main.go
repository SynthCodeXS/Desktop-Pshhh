package main

import (
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/gorilla/websocket"
)

type ChatMessange struct {
	Nick      string    `json:"nick"`
	Msg       string    `json:"msg"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	nick string

	messagesContainer *fyne.Container
	scrollContainer   *container.Scroll
)

func main() {

	fmt.Print("Введи свой ник: ")
	fmt.Scanln(&nick)

	//conn, err := net.Dial("tcp", "16.ip.gl.ply.gg:60348")
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		fmt.Println("net.Dial err:", err)
		os.Exit(1)
	}
	defer conn.Close()

	conn.WriteMessage(websocket.TextMessage, []byte(nick))

	myApp := app.New()
	myWindow := myApp.NewWindow("Test")
	myWindow.Resize(fyne.NewSize(400, 600))

	messagesContainer = container.NewVBox()
	scrollContainer = container.NewScroll(messagesContainer)

	inputEntry := widget.NewEntry()
	inputEntry.SetPlaceHolder("Send msg:")

	sendBtn := widget.NewButton("Send", func() {
		text := inputEntry.Text
		if text == "" {
			return
		}

		//fmt.Fprintf(conn, "%s\n", text)
		err := conn.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			fmt.Println("conn.WriteMessage err:", err)
		}
		inputEntry.SetText("")
	})

	//go func() {
	//	scanner := bufio.NewScanner(conn)
	//	for scanner.Scan() {
	//		var msg ChatMessange
	//		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
	//			continue
	//		}
	//		addLabel(msg)
	//	}
	//
	//}()
	go func() {
		for {
			var msg ChatMessange
			if err := conn.ReadJSON(&msg); err != nil {
				fmt.Println("conn.ReadJSON err:", err)
				return
			}

			addLabel(msg)

		}
	}()

	bottomArea := container.NewBorder(nil, nil, nil, sendBtn, inputEntry)
	content := container.NewBorder(nil, bottomArea, nil, nil, scrollContainer)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func addLabel(msg ChatMessange) {
	fyne.Do(func() {

		fullText := fmt.Sprintf("%s [%s]: %s",
			msg.Timestamp.Format("15:04"),
			msg.Nick,
			msg.Msg)

		l := widget.NewLabel(fullText)
		if msg.Nick == nick {
			l.Alignment = fyne.TextAlignTrailing
			l.TextStyle = fyne.TextStyle{Bold: true}
		} else {
			l.Alignment = fyne.TextAlignLeading
		}

		messagesContainer.Add(l)
		scrollContainer.ScrollToBottom()
		messagesContainer.Refresh()
	})
}
