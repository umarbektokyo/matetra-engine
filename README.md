# matetra v2.0.0
```
                         $$\                $$\                        
                         $$ |               $$ |                       
$$$$$$\$$$$\   $$$$$$\ $$$$$$\    $$$$$$\ $$$$$$\    $$$$$$\  $$$$$$\  
$$  _$$  _$$\  \____$$\\_$$  _|  $$  __$$\\_$$  _|  $$  __$$\ \____$$\ 
$$ / $$ / $$ | $$$$$$$ | $$ |    $$$$$$$$ | $$ |    $$ |  \__|$$$$$$$ |
$$ | $$ | $$ |$$  __$$ | $$ |$$\ $$   ____| $$ |$$\ $$ |     $$  __$$ |
$$ | $$ | $$ |\$$$$$$$ | \$$$$  |\$$$$$$$\  \$$$$  |$$ |     \$$$$$$$ |
\__| \__| \__| \_______|  \____/  \_______|  \____/ \__|      \_______|
```

A multiplayer math-based card game played in the terminal.

Players take turns defending their numbers while others attack. Use mathematical operations — addition, logarithms, factorials, theorems, and more — to grow your numbers past the limit or shrink your opponents'.

## Quick Start

### Server

```bash
# standalone
go build -o matetra-server ./cmd/matetra-server
./matetra-server start

# or with docker
docker compose up --build
```

The server listens on port `1729` by default (configurable via `PORT` env var).

### Client

```bash
go build -o matetra-client ./cmd/matetra-client
./matetra-client localhost:1729
```

## How to Play

### Setup

1. Each player connects and logs in with a username and password
2. One player creates a room (choosing win condition and hand size)
3. Others join using the 6-letter room code
4. Host starts the game — everyone gets cards and initial dice numbers

### Turn Flow

Each turn, one player is the **defender**. During their turn:

- **Everyone** can roll dice to fill empty number slots (press `d`)
- **Everyone** plays cards targeting the defender's numbers
  - The defender tries to **maximize** their own numbers
  - Attackers try to **minimize** the defender's numbers
  - Attackers sacrifice their own numbers as inputs (consumed after use)
- When all players press end turn (`e`), all queued cards resolve by precedence

### Card Types

| Type | Description |
|------|-------------|
| **Function** | Mathematical operations: add, subtract, multiply, divide, sqrt, log, sin, cos, etc. |
| **Constant** | Place a value into your set: pi, e, 42, fibonacci, etc. |
| **Theorem** | Special effects: swap numbers between any players, grant immunity, collapse adjacent numbers, decompose into primes, etc. |

### Win Conditions

- **Limit**: Numbers crossing the limit (starts at 1000) score points. First to 3 points wins. Limit multiplies by 10 each time someone scores.
- **Largest**: Biggest number after N turns wins.
- **Smallest**: Smallest number after N turns wins.

### Controls (TUI)

| Key | Action |
|-----|--------|
| `left/right` | Select card |
| `enter` | Play selected card |
| `d` | Roll dice |
| `e` | End turn |
| `f` | Finish dice phase |
| `:` | Command mode |
| `ctrl+c` | Quit |

When playing a card, the defender and dice values are auto-filled. You pick number slots (0-4). Some cards (like Googol, Commutative, Distributive) also ask you to pick a target player ID.

## Architecture

```
cmd/
  matetra-server/   Server binary
  matetra-client/   TUI client (Bubble Tea)
model/              Data types, number arithmetic (scientific notation)
engine/             Game logic, turn flow, win conditions
cards/              Card loading, dispatch
  constants/        Constant cards (pi, e, 42, ...)
  functions/        Function cards (add, sqrt, log, ...)
  theorems/         Theorem cards (swap, immunity, ...)
api/                HTTP + WebSocket server, JWT auth, rooms
utils/              Validation, hashing, helpers
```

### Number System

Numbers use scientific notation internally (`Value * 10^Base`) to handle very large and very small values without floating point overflow.

### Networking

- `POST /auth` — username + password, returns JWT token
- `GET /ws?token=...` — WebSocket connection (auto-rejoins existing game)
- All game actions happen over WebSocket messages
- Server broadcasts preview state (queued cards applied virtually) to all players
- Server sends WebSocket pings every 30s to keep connections alive through proxies/load balancers

## Configuration

Copy `.env.example` to `.env`:

```
JWT_SECRET=your-secret-key
PORT=1729
```

## Docker

```bash
# build and run
docker compose up --build

# or just build the image
docker build -t matetra .
docker run -p 1729:1729 -e JWT_SECRET=mysecret matetra
```

## Development

```bash
# run tests
go test ./...

# build both binaries
go build ./cmd/matetra-server
go build ./cmd/matetra-client

# vet
go vet ./...
```

## License

MIT
