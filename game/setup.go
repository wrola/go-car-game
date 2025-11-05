package game

import (
	"github.com/gorilla/websocket"
)

type PlayerConfig struct {
	ID         string
	X          float64
	Y          float64
	KeyMapping KeyMapping
}

type GameSetup struct {
	roomID        string
	playerConfigs []PlayerConfig
}
func NewGameSetup(roomID string) *GameSetup {
	return &GameSetup{
		roomID:        roomID,
		playerConfigs: make([]PlayerConfig, 0),
	}
}

func (gs *GameSetup) AddPlayer(config PlayerConfig) {
	gs.playerConfigs = append(gs.playerConfigs, config)
}
func (gs *GameSetup) Initialize(conn *websocket.Conn) (*Room, map[string]*InputHandler) {
	room := NewRoom(gs.roomID)

	inputHandlers := make(map[string]*InputHandler)
	for _, config := range gs.playerConfigs {
		player := NewPlayer(config.ID, conn, config.X, config.Y)

		room.AddPlayer(player)

		inputHandlers[config.ID] = NewInputHandler(config.KeyMapping)
	}

	return room, inputHandlers
}
