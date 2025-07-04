package main

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Azure/go-amqp"
	"golang.org/x/crypto/pkcs12"
)

func connect(url, p12Path, password string) error {
	data, err := os.ReadFile(p12Path)
	if err != nil {
		return fmt.Errorf("failed to read p12 file: %w", err)
	}
	blocks, err := pkcs12.ToPEM(data, password)
	if err != nil {
		return fmt.Errorf("failed to decode p12: %w", err)
	}
	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}
	cert, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		return fmt.Errorf("failed to parse key pair: %w", err)
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12,
	}
	_, err = amqp.Dial(context.Background(), url, &amqp.ConnOptions{TLSConfig: tlsConfig})
	if err != nil {
		return err
	}
	return nil
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

	connectBtn := widget.NewButton("Connect", func() {
		statusLabel.SetText("Connecting...")
		go func() {
			err := connect(urlEntry.Text, p12Entry.Text, passEntry.Text)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Connection failed: %v", err))
			} else {
				statusLabel.SetText("Connected successfully")
			}
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
	)

	w.SetContent(form)
	w.Resize(fyne.NewSize(400, 200))

	w.ShowAndRun()
}
