# AMQPS Client GUI

This project is an AMQP message browser written in Go. It uses the [Fyne](https://fyne.io/) toolkit to provide a simple graphical interface.

## Building

The project requires Go 1.19 or later.

```bash
# Install dependencies (requires internet access)
go mod tidy

# Build
 go build -o amqps-client
```

Running the binary will open a blank window. Future versions will allow connecting to an AMQP server and browsing messages.

