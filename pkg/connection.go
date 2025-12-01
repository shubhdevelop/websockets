package websocket

import (
	"fmt"
	"net"
)

type WsConnection struct {
	Conn           net.Conn // Raw TCP connection
	clientProtocol []string
}

func newWsConnection() WsConnection {
	return WsConnection{}
}

func (ws *WsConnection) HandleConnection() {
	defer ws.Conn.Close()
	for {
		message := ParseMessage(ws)
		if message == nil {
			continue
		}
		fmt.Println(string(message.ApplicationData))
	}
}
