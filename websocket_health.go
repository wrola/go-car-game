package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func startConnectionHealthcheck(conn *websocket.Conn) (done chan struct{}, cleanup func()) {
	const (
		pongWait   = 60 * time.Second
		pingPeriod = 30 * time.Second
		writeWait  = 10 * time.Second
	)

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	conn.SetReadDeadline(time.Now().Add(pongWait))

	pingTicker := time.NewTicker(pingPeriod)

	done = make(chan struct{})

	go func() {
		for {
			select {
			case <-pingTicker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("Error sending ping: %v", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	cleanup = func() {
		pingTicker.Stop()
	}

	return done, cleanup
}
