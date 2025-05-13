package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		panic("Dial error: " + err.Error())
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your name: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	conn.WriteMessage(websocket.TextMessage, []byte(username)) // Send username to server

	for {
		fmt.Println("\n==== Main Menu ====")
		fmt.Println("1. Send Message")
		fmt.Println("2. Live Feed")
		fmt.Println("3. Exit")
		fmt.Print("Choose option: ")

		var choice string
		fmt.Scan(&choice)

		switch choice {
		case "1":
			fmt.Print("Type your message: ")
			msg, _ := reader.ReadString('\n')
			msg = strings.TrimSpace(msg)
			conn.WriteMessage(websocket.TextMessage, []byte(msg))

		case "2":
			fmt.Println("Live messages (press Ctrl+C to stop):")
			go func() {
				for {
					_, msg, err := conn.ReadMessage()
					if err != nil {
						fmt.Println("Read error:", err)
						return
					}
					fmt.Println(string(msg))
				}
			}()

			var input string
			fmt.Scan(&input)
			if input == "q" {
				break
			}

		case "3":
			fmt.Println("Exiting...")
			leave_msg := fmt.Sprintf("%s leave from chat room", username)
			conn.WriteMessage(websocket.TextMessage, []byte(leave_msg))
			conn.Close()
			return

		default:
			fmt.Println("Invalid option. Try again.")
		}
	}
}
