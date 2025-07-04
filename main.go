package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Azure/go-amqp"
	"github.com/rivo/tview"
	"golang.org/x/crypto/pkcs12"
)

func connect(url, p12Path, password string) error {
	// Read the pkcs12 file
	data, err := ioutil.ReadFile(p12Path)
	if err != nil {
		return fmt.Errorf("failed to read p12 file: %w", err)
	}

	// Decode p12 to get certificate and key
	privKey, cert, err := pkcs12.Decode(data, password)
	if err != nil {
		return fmt.Errorf("failed to decode p12: %w", err)
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  privKey,
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	}

	_, err = amqp.Dial(context.Background(), url, &amqp.ConnOptions{TLSConfig: tlsConfig})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	app := tview.NewApplication()
	form := tview.NewForm()

	var url, p12Path, password string

	form.AddInputField("AMQPS URL", "", 40, nil, func(text string) { url = text })
	form.AddInputField("P12 Path", "", 40, nil, func(text string) { p12Path = text })
	form.AddPasswordField("P12 Password", "", 40, '*', func(text string) { password = text })

	status := tview.NewTextView().SetText("")

	form.AddButton("Connect", func() {
		status.SetText("Connecting...")
		if err := connect(url, p12Path, password); err != nil {
			status.SetText(fmt.Sprintf("Connection failed: %v", err))
		} else {
			status.SetText("Connected successfully")
		}
	})
	form.AddButton("Quit", func() {
		app.Stop()
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(status, 1, 0, false)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatalf("error running application: %v", err)
	}
}
