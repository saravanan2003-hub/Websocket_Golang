package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

type Message struct {
	Username string
	Content  string
}

var (
	user      = make(map[*websocket.Conn]string)
	broadcast = make(chan Message)
	mutex     sync.Mutex
	db        *sql.DB
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

	fmt.Println("DB instance is:", db)
	if db == nil {
		fmt.Println("DB is nil 🔥")
	}

	username := string(username_message)
	mutex.Lock()
	username_insert_query := "INSERT INTO user (username) VALUES (?)"
	_, err = db.Exec(username_insert_query, username)
	if err != nil {
		log.Fatal("Insert failed:", err)
	}
	user[conn] = username
	mutex.Unlock()

	join_msg := Message{
		Username: username,
		Content:  "Joined this room",
	}
	broadcast <- join_msg

	for {
		_, actual_msg, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(user, conn)
			mutex.Unlock()
			conn.Close()
			break
		}

		msg := Message{
			Username: username,
			Content:  string(actual_msg), // msgBytes is the message from client
		}
		broadcast <- msg

	}

}

func handleBroadcast() {
	for {
		msg := <-broadcast

		// 1. Store the incoming message in DB
		messageInsert := "INSERT INTO chat (username, message) VALUES (?, ?)"
		_, err := db.Exec(messageInsert, msg.Username, msg.Content)
		if err != nil {
			log.Println("Insert failed:", err)
			continue
		}

		msgTime := time.Now().Format("2006-01-02 15:04:05")
		finalMsg := fmt.Sprintf("%s -> %s : %s", msgTime, msg.Username, msg.Content)

		// 4. Send to all users
		mutex.Lock()
		for conn := range user {
			err := conn.WriteMessage(websocket.TextMessage, []byte(finalMsg))
			if err != nil {
				fmt.Println("Write error:", err)
				conn.Close()
				delete(user, conn)
			}
		}
		mutex.Unlock()
	}
}

func main() {

	dsn := "root:rootpass@tcp(127.0.0.1:3306)/chatApp"

	var err error                    // declare err only
	db, err = sql.Open("mysql", dsn) // assign to global db
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("DB not reachable:", err)
	}
	fmt.Println("✅ Database connected")

	http.HandleFunc("/ws", handleConnections)
	go handleBroadcast()

	fmt.Println("✅ Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
