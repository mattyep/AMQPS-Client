package main

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Azure/go-amqp"
	"golang.org/x/crypto/pkcs12"
)

var conn *amqp.Conn
var retrievedMsgs []*amqp.Message

func connect(url, p12Path, password string) (*amqp.Conn, error) {
	data, err := os.ReadFile(p12Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read p12 file: %w", err)
	}
	blocks, err := pkcs12.ToPEM(data, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decode p12: %w", err)
	}
	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}
	cert, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key pair: %w", err)
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12,
	}
	conn, err := amqp.Dial(context.Background(), url, &amqp.ConnOptions{TLSConfig: tlsConfig})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func main() {
	myApp := app.New()
	w := myApp.NewWindow("AMQPS Client")

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("amqps://hostname")

	p12Entry := widget.NewEntry()
	p12Entry.SetPlaceHolder("/path/to/client.p12")

	passEntry := widget.NewPasswordEntry()

	statusLabel := widget.NewLabel("")

	queueEntry := widget.NewEntry()
	queueEntry.SetPlaceHolder("queue name")

	numEntry := widget.NewEntry()
	numEntry.SetText("10")

	msgDisplay := widget.NewMultiLineEntry()
	msgDisplay.Disable()

	ackBtn := widget.NewButton("Acknowledge all", func() {
		go func() {
			for _, m := range retrievedMsgs {
				_ = m.Accept(context.Background())
			}
			retrievedMsgs = nil
			msgDisplay.SetText("")
		}()
	})
	ackBtn.Disable()

	releaseBtn := widget.NewButton("Release all", func() {
		go func() {
			for _, m := range retrievedMsgs {
				_ = m.Release(context.Background())
			}
			retrievedMsgs = nil
			msgDisplay.SetText("")
		}()
	})
	releaseBtn.Disable()

	connectBtn := widget.NewButton("Connect", func() {
		statusLabel.SetText("Connecting...")
		go func() {
			var err error
			conn, err = connect(urlEntry.Text, p12Entry.Text, passEntry.Text)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Connection failed: %v", err))
			} else {
				statusLabel.SetText("Connected successfully")
				ackBtn.Enable()
				releaseBtn.Enable()
			}
		}()
	})

	retrieveBtn := widget.NewButton("Retrieve", func() {
		if conn == nil {
			statusLabel.SetText("Not connected")
			return
		}
		num, err := strconv.Atoi(numEntry.Text)
		if err != nil {
			statusLabel.SetText("Invalid number")
			return
		}
		go func() {
			sess, err := conn.NewSession(context.Background(), nil)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Session error: %v", err))
				return
			}
			r, err := sess.NewReceiver(
				amqp.LinkSourceAddress(queueEntry.Text),
				amqp.LinkCredit(uint32(num)),
			)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Receiver error: %v", err))
				return
			}
			var lines []string
			for i := 0; i < num; i++ {
				msg, err := r.Receive(context.Background(), nil)
				if err != nil {
					statusLabel.SetText(fmt.Sprintf("Receive failed: %v", err))
					break
				}
				retrievedMsgs = append(retrievedMsgs, msg)
				lines = append(lines, string(msg.GetData()[0]))
			}
			msgDisplay.SetText(strings.Join(lines, "\n"))
		}()
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("AMQPS URL", urlEntry),
			widget.NewFormItem("P12 Path", p12Entry),
			widget.NewFormItem("P12 Password", passEntry),
		),
		connectBtn,
		statusLabel,
		widget.NewForm(
			widget.NewFormItem("Queue", queueEntry),
			widget.NewFormItem("Count", numEntry),
		),
		retrieveBtn,
		ackBtn,
		releaseBtn,
		msgDisplay,
	)

	w.SetContent(form)
	w.Resize(fyne.NewSize(400, 400))

	w.ShowAndRun()
}
