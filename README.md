# AMQPS Client

This application provides a small graphical interface for connecting to an AMQPS server using mutual TLS authentication.

## Features

- Text fields for the AMQPS URL, path to a `.p12` client certificate, and its password
- "Connect" button that attempts an mTLS connection using the `go-amqp` library
- Connection result is displayed inside the window
- The `.p12` file is converted to PEM at runtime to load the certificate chain

## Building

```
go build
```

## Running

```
go run .
```

A window will appear where you can enter the AMQPS server URL, the location of your `.p12` file and its password. Clicking **Connect** will attempt the connection and report success or any error.
