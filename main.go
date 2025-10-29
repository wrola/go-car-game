package main

import (
	"encoding/json"
	"log"
	"net/http"

	"go-car-game/game"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	staticFileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", staticFileServer)

	http.HandleFunc("/ws", handleLocalGame)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleLocalGame(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	room := game.NewRoom("local")

	player1 := game.NewPlayer("1", conn, 100.0, 200.0)
	player2 := game.NewPlayer("2", conn, 100.0, 350.0)

	room.AddPlayer(player1)
	room.AddPlayer(player2)

	log.Printf("Local multiplayer game started")

	connectionMessage := map[string]interface{}{
		"type":     "connected",
		"playerId": "local",
		"roomId":   room.ID,
		"mode":     "local",
	}
	conn.WriteJSON(connectionMessage)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		var clientMessage map[string]interface{}
		if err := json.Unmarshal(message, &clientMessage); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		messageType, ok := clientMessage["type"].(string)
		if !ok {
			continue
		}

		switch messageType {
		case "input":
			if inputData, ok := clientMessage["input"].(map[string]interface{}); ok {
				player1InputState := make(map[string]bool)
				if val, ok := inputData["w"].(bool); ok {
					player1InputState["up"] = val
				}
				if val, ok := inputData["s"].(bool); ok {
					player1InputState["down"] = val
				}
				if val, ok := inputData["a"].(bool); ok {
					player1InputState["left"] = val
				}
				if val, ok := inputData["d"].(bool); ok {
					player1InputState["right"] = val
				}
				room.HandlePlayerInput("1", player1InputState)

				player2InputState := make(map[string]bool)
				if val, ok := inputData["up"].(bool); ok {
					player2InputState["up"] = val
				}
				if val, ok := inputData["down"].(bool); ok {
					player2InputState["down"] = val
				}
				if val, ok := inputData["left"].(bool); ok {
					player2InputState["left"] = val
				}
				if val, ok := inputData["right"].(bool); ok {
					player2InputState["right"] = val
				}
				room.HandlePlayerInput("2", player2InputState)
			}
		}
	}

	log.Printf("Local game ended")
}
