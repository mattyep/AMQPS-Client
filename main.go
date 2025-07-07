package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/Azure/go-amqp"
)

var conn *amqp.Conn
var retrievedMsgs []*amqp.Message
var receiver *amqp.Receiver

func connect(url, username, password string) (*amqp.Conn, error) {
	conn, err := amqp.Dial(context.Background(), url, &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain(username, password),
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func retrieveMessages(ctx context.Context, conn *amqp.Conn, queueName string, nbMsg int) ([]string, error) {
	sess, err := conn.NewSession(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("session error: %v", err)
	}
	if receiver != nil {
		receiver.Close(ctx)
	}

	receiver, err = sess.NewReceiver(
		ctx,
		queueName,
		&amqp.ReceiverOptions{
			Capabilities:       []string{"queue"},
			SourceCapabilities: []string{"queue"},
			Credit:             10000,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("receiver error: %v", err)
	}
	var lines []string
	for i := 0; i < nbMsg; i++ {
		msg, err := receiver.Receive(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("receive failed: %v", err)
		}

		var mappedMsg map[string]any
		_ = json.Unmarshal(msg.GetData(), &mappedMsg)
		jsonIndent, _ := json.MarshalIndent(mappedMsg, "", "\t")
		_ = receiver.ReleaseMessage(ctx, msg)
		retrievedMsgs = append(retrievedMsgs, msg)
		lines = append(lines, string(jsonIndent))
	}
	return lines, nil
}

func createInputField(placeholder, text string) *widget.Entry {
	field := widget.NewEntry()
	field.SetPlaceHolder(placeholder)
	field.SetText(text)
	return field
}

func topBar() *fyne.Container {
	appName := canvas.NewText("AMQP Client", color.White)
	appName.TextStyle = fyne.TextStyle{Bold: true}
	line := canvas.NewLine(color.RGBA{255, 255, 255, 50})
	line.StrokeWidth = 1
	top := container.NewVBox(appName, line)
	return top
}

func main() {
	myApp := app.New()
	w := myApp.NewWindow("AMQPS Client")
	w.Resize(fyne.NewSize(1440, 900))

	urlField := createInputField("amqps://hostname", "amqp://localhost:5672")
	usernameField := createInputField("username", "user")
	passEntry := createInputField("password", "user")
	statusLabel := widget.NewLabel("")
	queueNameField := createInputField("queue name", "my_queue")
	nbMsgField := createInputField("count", "10")

	data := binding.BindStringList(
		&[]string{},
	)

	msgList := widget.NewListWithData(data,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	connectBtn := widget.NewButton("Connect", func() {
		statusLabel.SetText("Connecting...")
		go func() {
			var err error
			conn, err = connect(urlField.Text, usernameField.Text, passEntry.Text)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Connection failed: %v", err))
			} else {
				statusLabel.SetText("Connected successfully")
			}
		}()
	})

	retrieveBtn := widget.NewButton("Retrieve", func() {
		if conn == nil {
			statusLabel.SetText("Not connected")
			return
		}
		nbMsg, err := strconv.Atoi(nbMsgField.Text)
		if err != nil {
			statusLabel.SetText("Invalid number")
			return
		}
		go func() {
			lines, err := retrieveMessages(context.Background(), conn, queueNameField.Text, nbMsg)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Retrieve failed: %v", err))
				return
			}
			for _, line := range lines {
				data.Append(line)
			}
		}()
	})

	mainView := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("AMQPS URL", urlField),
			widget.NewFormItem("Username", usernameField),
			widget.NewFormItem("Password", passEntry),
		),
		connectBtn,
		statusLabel,
		widget.NewForm(
			widget.NewFormItem("Queue", queueNameField),
			widget.NewFormItem("Count", nbMsgField),
		),
		retrieveBtn,
		msgList,
	)

	leftMenu := container.NewVBox(canvas.NewText("Left Menu", color.White))
	content := container.NewBorder(topBar(), nil, leftMenu, nil, mainView)
	w.SetContent(content)

	w.ShowAndRun()
}
