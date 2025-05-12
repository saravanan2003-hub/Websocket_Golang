package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	user      = make(map[*websocket.Conn]string)
	broadcast = make(chan string)
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // allow all origins (for local testing)
		},
	}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}

	_, username_message, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Username read error:", err)
		conn.Close()
		return
	}

	username := string(username_message)
	user[conn] = username

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			delete(user, conn)
			conn.Close()
			break
		}

		final := fmt.Sprintf("%s: %s", username, string(msg))
		broadcast <- final

	}

}

func handleBroadcast() {
	for {
		msg := <-broadcast
		for conn := range user {
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				fmt.Println("Write error:", err)
				conn.Close()
				delete(user, conn)
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	go handleBroadcast()

	fmt.Println("✅ Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
