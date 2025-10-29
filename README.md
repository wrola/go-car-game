# Go Racing Game - Local Multiplayer

A simple 2D racing game for two players built with Go and WebSockets. Both players play on the same browser using different controls.

## Features

- **Local Multiplayer**: Two players race on the same screen
- **Real-time Physics**: Car acceleration, friction, and steering
- **Complex Track**: 6 checkpoints to navigate through
- **Simple Architecture**: Clean, easy-to-understand codebase

## Controls

- **Player 1 (Red Car)**: W/A/S/D
  - W: Accelerate
  - S: Brake/Reverse
  - A: Turn Left
  - D: Turn Right

- **Player 2 (Cyan Car)**: Arrow Keys
  - ↑: Accelerate
  - ↓: Brake/Reverse
  - ←: Turn Left
  - →: Turn Right

## Quick Start

1. **Install Go** (version 1.24+)

2. **Clone and run:**
   ```bash
   cd go-game
   go run main.go
   ```

3. **Open your browser:**
   ```
   http://localhost:8080
   ```

4. **Race!** First player to pass all 6 checkpoints wins!

## Project Structure

```
go-game/
├── main.go              # HTTP server and WebSocket handler
├── game/
│   ├── player.go        # Player physics and input handling
│   └── room.go          # Game loop, checkpoints, win detection
├── static/
│   ├── index.html       # Game UI
│   ├── game.js          # Client-side rendering and controls
│   └── style.css        # Styling
└── blog-post-locks.md   # Educational content about mutex locks
```

## Architecture

This game demonstrates Go concurrency with a simple architecture:

### Goroutines (3 total):
1. **Main goroutine** - HTTP server
2. **Input goroutine** - WebSocket handler (both players)
3. **Game loop goroutine** - Updates positions at 20 FPS

### Why Locks Are Needed:
- Input goroutine writes player speeds when keys are pressed
- Game loop goroutine reads/writes positions and speeds
- Without locks → race conditions (lost inputs, corrupted positions)
- With locks → safe concurrent access

## Technical Details

- **WebSocket**: Real-time bidirectional communication
- **Game Loop**: 50ms tick rate (20 FPS)
- **Physics**: Realistic car movement with friction and steering
- **Checkpoints**: Must be passed in order to finish

## License

MIT License - Feel free to learn from and modify this code!
