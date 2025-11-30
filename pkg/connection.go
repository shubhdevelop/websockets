package websocket

import "net"

type WsConnection struct {
	Conn           net.Conn // Raw TCP connection
	clientProtocol []string
}

func newWsConnection() WsConnection {
	return WsConnection{}
}
