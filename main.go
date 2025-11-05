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

	setup := game.NewGameSetup("local")
	setup.AddPlayer(game.PlayerConfig{
		ID:         "1",
		X:          100.0,
		Y:          200.0,
		KeyMapping: game.WASDMapping,
	})
	setup.AddPlayer(game.PlayerConfig{
		ID:         "2",
		X:          100.0,
		Y:          350.0,
		KeyMapping: game.ArrowKeyMapping,
	})

	room, inputHandlers := setup.Initialize(conn)

	log.Printf("Local multiplayer game started")

	connectionMessage := map[string]interface{}{
		"type":     "connected",
		"playerId": "local",
		"roomId":   room.ID,
		"mode":     "local",
	}
	conn.WriteJSON(connectionMessage)

	done, cleanup := startConnectionHealthcheck(conn)
	defer cleanup()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			close(done)
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
				for playerID, handler := range inputHandlers {
					movementState := handler.Translate(inputData)
					room.HandlePlayerInput(playerID, movementState)
				}
			}
		}
	}

	room.MarkConnectionClosed()
	room.TriggerShutdown()

	log.Printf("Local game ended")
}
