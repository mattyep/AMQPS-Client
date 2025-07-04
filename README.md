# AMQPS Client

This is a simple terminal UI application that connects to an AMQPS server using a client certificate in a PKCS#12 (.p12) file.

## Features

- Input fields for the AMQPS URL, the path to a `.p12` file, and the password protecting that file.
- Connect button that establishes an AMQPS connection using mutual TLS.
- Displays whether the connection succeeded or if an error occurred.

## Building

```
go build
```

## Running

Run the resulting binary or use `go run`:

```
go run .
```

You will be presented with a form to provide the AMQPS URL, path to the `.p12` file, and its password. After pressing **Connect** the application reports the result.

