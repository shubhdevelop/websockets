package websocket

type Websocket struct {
	Connections map[*WsConnection]bool
	Broadcast   chan []byte
	register    chan *WsConnection
	unregister  chan *WsConnection
}
