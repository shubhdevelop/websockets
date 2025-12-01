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
		go conn.HandleConnection()
	})

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
