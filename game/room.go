package game

import (
	"encoding/json"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Checkpoint struct {
	X      float64
	Y      float64
	Radius float64
}

type Room struct {
	ID                string
	Players           []*Player
	Started           bool
	Winner            string
	Checkpoints       []Checkpoint
	stateMutex        sync.RWMutex
	gameLoopTicker    *time.Ticker
	gameStopChannel   chan bool
	InactivityTimeout time.Duration
	connectionClosed  int32 // atomic: 0 = open, 1 = closed
}

func NewRoom(id string) *Room {
	checkpoints := []Checkpoint{
		{X: 200, Y: 300, Radius: 50},
		{X: 400, Y: 450, Radius: 50},
		{X: 600, Y: 300, Radius: 50},
		{X: 800, Y: 450, Radius: 50},
		{X: 1000, Y: 300, Radius: 50},
		{X: 1200, Y: 200, Radius: 50},
	}

	return &Room{
		ID:                id,
		Players:           make([]*Player, 0, 2),
		Started:           false,
		Checkpoints:       checkpoints,
		gameStopChannel:   make(chan bool),
		InactivityTimeout: 30 * time.Second,
	}
}

func (r *Room) AddPlayer(player *Player) bool {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	if len(r.Players) >= 2 {
		return false
	}

	r.Players = append(r.Players, player)

	if len(r.Players) == 2 {
		r.Started = true
		go r.StartGameLoop()
	}

	return true
}

func (r *Room) RemovePlayer(playerID string) {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	for i, player := range r.Players {
		if player.ID == playerID {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			break
		}
	}

	if r.Started {
		r.gameStopChannel <- true
	}
}

func (r *Room) StartGameLoop() {
	r.gameLoopTicker = time.NewTicker(50 * time.Millisecond)
	defer r.gameLoopTicker.Stop()

	inactivityCheckTicker := time.NewTicker(5 * time.Second)
	defer inactivityCheckTicker.Stop()

	log.Printf("Game started in room %s", r.ID)

	for {
		select {
		case <-r.gameLoopTicker.C:
			// Skip update if connection is closed
			if r.isConnectionClosed() {
				log.Printf("Connection closed, stopping game loop in room %s", r.ID)
				return
			}
			r.Update()
			r.BroadcastGameState()
		case <-inactivityCheckTicker.C:
			r.CheckInactivePlayers()
		case <-r.gameStopChannel:
			log.Printf("Game stopped in room %s", r.ID)
			return
		}
	}
}

func (r *Room) Update() {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	if r.Winner != "" {
		return
	}

	for _, player := range r.Players {
		player.UpdatePosition()

		if player.Checkpoint < len(r.Checkpoints) {
			currentCheckpoint := r.Checkpoints[player.Checkpoint]
			deltaX := player.PositionX - currentCheckpoint.X
			deltaY := player.PositionY - currentCheckpoint.Y
			distance := math.Sqrt(deltaX*deltaX + deltaY*deltaY)

			if distance < currentCheckpoint.Radius {
				player.Checkpoint++
				log.Printf("Player %s reached checkpoint %d", player.ID, player.Checkpoint)

				if player.Checkpoint >= len(r.Checkpoints) && !player.Finished {
					player.Finished = true
					if r.Winner == "" {
						r.Winner = player.ID
						log.Printf("Player %s won the race!", player.ID)
					}
				}
			}
		}
	}
}

func (r *Room) CheckInactivePlayers() {
	r.stateMutex.RLock()
	inactivePlayers := make([]string, 0)

	for _, player := range r.Players {
		if player.IsInactive(r.InactivityTimeout) {
			inactivePlayers = append(inactivePlayers, player.ID)
		}
	}
	r.stateMutex.RUnlock()

	if len(inactivePlayers) > 0 {
		for _, playerID := range inactivePlayers {
			log.Printf("Player %s inactive for %v, closing connection and shutting down", playerID, r.InactivityTimeout)
			r.stateMutex.RLock()
			for _, player := range r.Players {
				if player.ID == playerID {
					player.Conn.Close()
					break
				}
			}
			r.stateMutex.RUnlock()
		}
		r.markConnectionClosed()
		r.triggerShutdown()
	}
}

func (r *Room) BroadcastGameState() {
	if r.isConnectionClosed() {
		return
	}

	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()

	gameState := r.SerializeRoomState()
	serializedData, err := json.Marshal(gameState)
	if err != nil {
		log.Printf("Error marshaling state: %v", err)
		return
	}

	for _, player := range r.Players {
		err := player.Conn.WriteMessage(websocket.TextMessage, serializedData)
		if err != nil {
			log.Printf("Error sending state to player %s: %v. Marking connection as closed.", player.ID, err)
			r.markConnectionClosed()
			r.triggerShutdown()
			return
		}
	}
}

func (r *Room) SerializeRoomState() map[string]interface{} {
	playerStates := make([]map[string]interface{}, 0)
	for _, player := range r.Players {
		playerStates = append(playerStates, player.SerializePlayerState())
	}

	checkpointStates := make([]map[string]interface{}, 0)
	for _, checkpoint := range r.Checkpoints {
		checkpointStates = append(checkpointStates, map[string]interface{}{
			"x":      checkpoint.X,
			"y":      checkpoint.Y,
			"radius": checkpoint.Radius,
		})
	}

	return map[string]interface{}{
		"type":        "gameState",
		"players":     playerStates,
		"winner":      r.Winner,
		"started":     r.Started,
		"checkpoints": checkpointStates,
	}
}

func (r *Room) IsFull() bool {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	return len(r.Players) >= 2
}

func (r *Room) HandlePlayerInput(playerID string, input map[string]bool) {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()

	for _, player := range r.Players {
		if player.ID == playerID {
			player.ProcessMovementInput(input["up"], input["down"], input["left"], input["right"])
			break
		}
	}
}

func (r *Room) isConnectionClosed() bool {
	return atomic.LoadInt32(&r.connectionClosed) == 1
}

func (r *Room) MarkConnectionClosed() {
	atomic.StoreInt32(&r.connectionClosed, 1)
}

func (r *Room) markConnectionClosed() {
	r.MarkConnectionClosed()
}

func (r *Room) TriggerShutdown() {
	if r.Started {
		select {
		case r.gameStopChannel <- true:
		default:
		}
	}
}

func (r *Room) triggerShutdown() {
	r.TriggerShutdown()
}
