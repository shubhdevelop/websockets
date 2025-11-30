package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	magic_string = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

var Origin map[string]bool = map[string]bool{
	"*": true,
}

type WsUpgrade struct {
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*WsConnection, error) {
	wsConn := newWsConnection()
	var answerKey string
	if r.Method != http.MethodGet {
		return nil, errors.New("Bad handshake METHOD not GET")
	}
	if r.Header.Get("Upgrade") != "websocket" {
		return nil, errors.New("Bad handshake: Wrong Upgrade Value")
	}
	if r.Header.Get("Connection") != "Upgrade" {
		return nil, errors.New("Bad handshake: Connection type not Upgrade")
	}
	if !Origin["*"] {
		return nil, errors.New("Bad handshake: Origin Not Allowed")
	}
	if protocols := r.Header.Get("Sec-WebSocket-Protocol"); protocols != "" {
		wsConn.clientProtocol = strings.Split(strings.TrimSpace(protocols), ",")
		for i := range wsConn.clientProtocol {
			wsConn.clientProtocol[i] = strings.TrimSpace(wsConn.clientProtocol[i])
		}
	}

	if r.Header.Get("Sec-WebSocket-Key") == "" {
		return nil, errors.New("Bad handshake: Websocket key must be present")
	} else {
		answerKey = computeAcceptKey(strings.TrimSpace(r.Header.Get("Sec-WebSocket-key")))
	}
	if r.Header.Get("Sec-WebSocket-Version") != "13" {
		return nil, errors.New("Bad handshake: Websocket Protocol version must be 13")
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("Hijacking not supported")
	}
	conn, brw, err := hijacker.Hijack()
	if err != nil {
		return nil, errors.New("Hijacking not supported")
	}
	wsConn.Conn = conn

	fmt.Fprintf(brw,
		"HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n\r\n",
		answerKey,
	)
	brw.Flush()

	return &wsConn, nil
}

func computeAcceptKey(key string) string {
	sum := sha1.Sum([]byte(key + magic_string))
	return base64.StdEncoding.EncodeToString(sum[:])
}
