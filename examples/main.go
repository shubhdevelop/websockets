package main

import (
	"fmt"
	"log"
	"net/http"

	websocket "github.com/shubhdevelop/websockets/pkg"
)

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Upgrade(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("upgrade error: %v", err)
			return
		}
		fmt.Println("WebSocket handshake complete")

		for {
			frame, err := websocket.ParseNetworkFrame(conn)
			fmt.Println(string(frame.PayloadData))
			if err != nil {
				fmt.Println("client disconnected")
				return
			}
			if frame.Opcode == 0x09 {
				frame := websocket.NewPongFrame("reply to ping")
				conn.Conn.Write(frame.ComponseNetworkFrame())
				continue
			}
			frame.Mask = false
			fmt.Println(string(frame.PayloadData))
			conn.Conn.Write(frame.ComponseNetworkFrame())
		}
	})

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
