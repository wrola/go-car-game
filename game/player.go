package game

import (
	"github.com/gorilla/websocket"
	"math"
	"sync"
	"time"
)

const (
	CarAcceleration = 1.0
	CarMaxSpeed     = 12.0
	CarReverseSpeed = -6.0
	CarTurnSpeed    = 5.0
	SpeedDecayRate  = 0.99
	MinimumSpeed    = 0.2
)

type Player struct {
	ID               string
	Conn             *websocket.Conn
	PositionX        float64
	PositionY        float64
	Angle            float64
	Speed            float64
	Finished         bool
	Checkpoint       int
	LastActivity     time.Time
	stateMutex       sync.Mutex
}

func NewPlayer(id string, conn *websocket.Conn, initialX, initialY float64) *Player {
	return &Player{
		ID:           id,
		Conn:         conn,
		PositionX:    initialX,
		PositionY:    initialY,
		Angle:        0,
		Speed:        0,
		Finished:     false,
		LastActivity: time.Now(),
	}
}

func (p *Player) UpdatePosition() {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	angleInRadians := p.Angle * (math.Pi / 180)
	p.PositionX += math.Cos(angleInRadians) * p.Speed
	p.PositionY += math.Sin(angleInRadians) * p.Speed

	p.Speed *= SpeedDecayRate
	if math.Abs(p.Speed) < MinimumSpeed {
		p.Speed = 0
	}
}

func (p *Player) ProcessMovementInput(up, down, left, right bool) {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if up || down || left || right {
		p.LastActivity = time.Now()
	}

	if up {
		p.Speed += CarAcceleration
		if p.Speed > CarMaxSpeed {
			p.Speed = CarMaxSpeed
		}
	}
	if down {
		p.Speed -= CarAcceleration
		if p.Speed < CarReverseSpeed {
			p.Speed = CarReverseSpeed
		}
	}

	if math.Abs(p.Speed) > 0.5 {
		if left {
			p.Angle -= CarTurnSpeed
		}
		if right {
			p.Angle += CarTurnSpeed
		}
	}

	if p.Angle < 0 {
		p.Angle += 360
	}
	if p.Angle >= 360 {
		p.Angle -= 360
	}
}

func (p *Player) IsInactive(timeout time.Duration) bool {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()
	return time.Since(p.LastActivity) > timeout
}

func (p *Player) SerializePlayerState() map[string]interface{} {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	return map[string]interface{}{
		"id":         p.ID,
		"x":          p.PositionX,
		"y":          p.PositionY,
		"angle":      p.Angle,
		"speed":      p.Speed,
		"finished":   p.Finished,
		"checkpoint": p.Checkpoint,
	}
}
