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
			b := make([]byte, 2)
			_, err := conn.Conn.Read(b)
			if err != nil {
				fmt.Println("client disconnected")
				return
			}
			frame := websocket.NewFrame(b)
			_, err = conn.Conn.Write(frame.ComponseNetworkFrame())
			if err != nil {
				fmt.Println("error writing to client", err)
				return
			}
			fmt.Printf("received raw bytes: % x\n", b)
		}
	})

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
