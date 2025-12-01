# Copilot Instructions for websockets

This repository contains a Go implementation of the WebSocket protocol (RFC 6455).

## Project Overview

This is a lightweight WebSocket library for Go that provides:
- WebSocket server-side connection handling via HTTP upgrade
- WebSocket frame parsing and composing
- Support for text, binary, ping, pong, and close frames
- Example server application with a web frontend for testing

## Code Structure

- `pkg/` - Main WebSocket library package
  - `server.go` - HTTP upgrade handling and WebSocket handshake
  - `connection.go` - WebSocket connection wrapper
  - `frame.go` - WebSocket frame parsing and composing
  - `constants.go` - WebSocket protocol constants (close codes, opcodes)
  - `client.go` - Client-side functionality (placeholder)
  - `utils.go` - Utility functions (placeholder)
- `examples/` - Example applications
  - `main.go` - Example WebSocket echo server
  - `frontend/` - HTML test client

## Development Guidelines

### Go Conventions
- Use Go idioms and follow the standard Go project layout
- Run `go fmt` to format code before committing
- Run `go vet` to check for common errors
- Run `go test ./...` to execute tests

### WebSocket Protocol
- This library implements RFC 6455 WebSocket protocol
- Frame opcodes: 0x0 (continuation), 0x1 (text), 0x2 (binary), 0x8 (close), 0x9 (ping), 0xA (pong)
- Server-to-client frames are unmasked; client-to-server frames are masked

### Testing
- Tests are located alongside source files with `_test.go` suffix
- Use table-driven tests where appropriate
- Test both happy path and error cases

### Building and Running

Build the example server:
```bash
go build -o server ./examples
```

Run the example server:
```bash
go run ./examples
```

The server listens on port 8080 and serves WebSocket connections at `/ws`.

## Important Notes

- The package name is `websocket` (in the `pkg` directory)
- Import path: `github.com/shubhdevelop/websockets/pkg`
- This project is licensed under GPL-3.0
