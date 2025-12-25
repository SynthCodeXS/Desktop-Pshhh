package main

import (
	"fmt"
	"image/color"
	"os"
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
	Nick      string    `json:"nick"`
	To        string    `json:"to"`
	Msg       string    `json:"msg"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	nick string

	messagesContainer *fyne.Container
	scrollContainer   *container.Scroll
)

func main() {

	fmt.Print("Send username: ")
	fmt.Scanln(&nick)

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8443/ws", nil)
	if err != nil {
		fmt.Println("net.Dial err:", err)
		os.Exit(1)
	}
	defer conn.Close()

	conn.WriteMessage(websocket.TextMessage, []byte(nick))

	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())

	myWindow := myApp.NewWindow("Test")
	myWindow.Resize(fyne.NewSize(400, 600))

	messagesContainer = container.NewVBox()
	scrollContainer = container.NewScroll(messagesContainer)

	inputEntry := widget.NewEntry()
	inputEntry.SetPlaceHolder("Send msg:")
	targetEntry := widget.NewEntry()
	targetEntry.SetPlaceHolder("Target:")

	inputContainer := container.NewGridWithColumns(2, targetEntry, inputEntry)

	sendBtn := widget.NewButton("Send", func() {
		text := inputEntry.Text
		if text == "" {
			return
		}
		targetText := targetEntry.Text
		if text == "" {
			return
		}

		msg := ChatMessage{
			Msg: text,
			To:  targetText,
		}

		err := conn.WriteJSON(msg)
		if err != nil {
			fmt.Println("conn.WriteMessage err:", err)
		}
		inputEntry.SetText("")
	})

	//go func() {
	//	scanner := bufio.NewScanner(conn)
	//	for scanner.Scan() {
	//		var msg ChatMessage
	//		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
	//			continue
	//		}
	//		addLabel(msg)
	//	}
	//
	//}()
	go func() {
		for {
			var msg ChatMessage
			if err := conn.ReadJSON(&msg); err != nil {
				fmt.Println("conn.ReadJSON err:", err)
				return
			}

			addLabel(msg)

		}
	}()

	bottomArea := container.NewBorder(nil, nil, nil, sendBtn, inputContainer)
	content := container.NewBorder(nil, bottomArea, nil, nil, scrollContainer)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
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
				headerText = fmt.Sprintf("🔒 %s (V)", msg.Nick)
			}
		} else {
			if isMine {
				bubbleColor = color.NRGBA{R: 30, G: 60, B: 150, A: 255}
				headerText = "You"
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

		//bgColor := color.NRGBA{R: 20, G: 30, B: 80, A: 255}
		//background := canvas.NewRectangle(bgColor)

		//l := widget.NewLabel(msg.Msg)
		//l.Wrapping = fyne.TextWrapWord
		//l.TextStyle.Bold = true
		//
		//if msg.Nick == nick {
		//	l.Alignment = fyne.TextAlignTrailing
		//} else {
		//	l.Alignment = fyne.TextAlignLeading
		//}
		//
		//timeL := widget.NewLabel(msg.Timestamp.Format("15:04"))
		//timeL.Alignment = fyne.TextAlignTrailing
		//
		//contentBox := container.NewVBox(
		//	l,
		//	timeL,
		//)
		//
		////statusBar := container.NewStack(background, container.NewPadded(contentBox))
		//
		////card := widget.NewCard(msg.Nick+":"+msg.Timestamp.Format("15:04"), "", l)
		////
		////if msg.Nick == nick {
		////	box := container.NewHBox(layout.NewSpacer(), card)
		////	messagesContainer.Add(box)
		////} else {
		////	box := container.NewHBox(card, layout.NewSpacer())
		////	messagesContainer.Add(box)
		////}
		//
		////if msg.Nick == nick {
		////	l.Alignment = fyne.TextAlignTrailing
		////	l.TextStyle = fyne.TextStyle{Bold: true}
		////} else {
		////	l.Alignment = fyne.TextAlignLeading
		////}
		//messagesContainer.Add(contentBox)
		//scrollContainer.ScrollToBottom()
		//messagesContainer.Refresh()
	})
}
