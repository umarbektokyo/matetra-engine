# Matetra: A Multiplayer Mathematical Card Game Engine

## Abstract

Matetra is a real-time multiplayer card game where players compete by applying mathematical operations to numerical values. Each player maintains a set of 5 numbers, and through strategic use of function cards, constant cards, and theorem cards, they attempt to grow their numbers past a scoring threshold or manipulate opponents' numbers to their disadvantage. The game combines elements of traditional card games with mathematical reasoning, creating a competitive environment where arithmetic, algebra, and number theory become weapons.

This document describes the complete game design, the internal systems that power it, and the architecture that enables real-time multiplayer play.

---

## 1. Core Concept

### 1.1 The Premise

Every player has a **set of 5 number slots**. Numbers live in these slots and can be manipulated through cards. The game proceeds in rounds called turns. Each turn, one player is designated the **defender** — everyone else attacks their numbers while the defender tries to protect or grow them.

The central tension: attackers want to shrink the defender's numbers (or make them negative, zero, or otherwise useless), while the defender wants to grow them past a scoring limit.

### 1.2 The Goal

Depending on the chosen win condition:

- **Limit mode** (default): Any number that crosses the current limit (starts at 1,000) scores a point for its owner. The number is consumed and the slot becomes empty. First player to 3 points wins. Each time anyone scores, the limit multiplies by 10 (1,000 → 10,000 → 100,000 → ...).
- **Largest mode**: After a fixed number of turns, the player with the single largest number wins.
- **Smallest mode**: After a fixed number of turns, the player with the single smallest number wins.

---

## 2. The Number System

### 2.1 Scientific Notation Representation

Matetra needs to handle numbers ranging from extremely small to astronomically large. A naive floating-point approach would overflow or lose precision. Instead, every number is stored in **scientific notation**:

```
Number = Value x 10^Base
```

Where:
- **Value** is a float64 coefficient, normalized to the range [1, 10) (or negative equivalent)
- **Base** is an int64 exponent

Examples:
| Real Value | Value | Base | Display |
|------------|-------|------|---------|
| 3,140 | 3.14 | 3 | 3140 |
| 0.005 | 5.0 | -3 | 0.005 |
| 6.02 x 10^23 | 6.02 | 23 | 6.02e23 |
| -42 | -4.2 | 1 | -42 |

This representation allows arithmetic on numbers of wildly different magnitudes without overflow. Squaring a number doubles its base. Taking a square root halves it. Multiplication adds bases. Division subtracts them.

### 2.2 Arithmetic Operations

All arithmetic operates on the scientific notation directly:

- **Addition/Subtraction**: Align bases, then add/subtract values. Renormalize.
- **Multiplication**: Multiply values, add bases. Renormalize.
- **Division**: Divide values, subtract bases. Renormalize.
- **Square Root**: If the base is odd, multiply value by 10 and decrement base to make it even. Then sqrt the value and halve the base.
- **Exponentiation**: For very large exponents (e^x where x > 709), convert to base-10 representation to avoid float64 overflow.

Every operation ends with **normalization** (adjusting value back into [1, 10) range) and **sanitization** (rejecting NaN or Infinity results).

### 2.3 Number Marks

Each number slot carries a **mark** that determines its special status:

| Mark | Meaning | Behavior |
|------|---------|----------|
| `""` | Normal | Can be targeted by any card |
| `"n"` | Null / Empty | Slot is empty. Can be filled by dice or constants |
| `"I"` | Immune | Protected from all card effects for one turn. Immunity expires at the start of the next turn |
| `"F"` | Fibonacci | Automatically grows to the next Fibonacci number each turn |
| `"FI"` | Fibonacci + Immune | Both protections active. Reverts to `"F"` when immunity expires |

---

## 3. Cards

The deck contains **49 cards** across three types. Each card has a **precedence** value that determines resolution order when multiple cards are queued.

### 3.1 Function Cards (24 cards)

Function cards perform mathematical operations. They typically target the **defender's** numbers and consume the attacker's numbers as operands.

#### Binary Operations (4 copies each)
These take two numbers — a **target** (on the defender's set) and a **source** (from the attacker's set). The source is consumed (becomes null) after use.

| Card | Formula | Notes |
|------|---------|-------|
| Addition | target = target + source | |
| Subtraction | target = target - source | |
| Multiplication | target = target x source | |
| Division | target = target / source | Cannot divide by zero |

#### Unary Operations (1-2 copies each)
These modify a single number on the defender's set.

| Card | Formula | Notes |
|------|---------|-------|
| Absolute Value | target = \|target\| | |
| Inverse | target = 1 / target | Cannot invert zero |
| Negative | target = -target | |
| Positive | target = +target | Does nothing (it's a joke card) |
| Square | target = target^2 | |
| Square Root | target = sqrt(target) | Must be non-negative |
| Base-10 Logarithm | target = log10(target) | Must be positive |
| Natural Logarithm | target = ln(target) | Must be positive |
| Exponential | target = e^target | Handles overflow via log conversion |

#### Dice-Modified Operations (1 copy each)
These involve a random dice roll (1-6) as a modifier.

| Card | Formula | Notes |
|------|---------|-------|
| Sine | target = target x sin(d) | d in radians |
| Cosine | target = target x cos(d) | d in radians |
| Tangent | target = target x tan(d) | d in radians; undefined cases rejected |
| Logarithm | target = log_d(target) | Base-d logarithm; base 1 rejected |
| Base-Root | target = target^(1/d) | d-th root |
| Power | target = target^d | |

#### Aggregate Operations (1 copy each)
These operate on an **entire player's set** at once.

| Card | Formula | Notes |
|------|---------|-------|
| Summation (Sigma) | Sums all numbers into one slot, others become empty | Immune numbers are excluded |
| Product (Pi) | Multiplies all numbers into one slot, others become empty | Immune numbers are excluded |

#### Polynomial Operations (1 copy each)
These combine multiple numbers with a dice roll.

| Card | Formula | Notes |
|------|---------|-------|
| First Order | target = target x d + source | Consumes source |
| Second Order | target = target x d^2 + source1 x d + source2 | Consumes both sources |

### 3.2 Constant Cards (17 cards)

Constant cards place a value into the player's own set. They fill the first empty slot, or replace the smallest non-immune number if all slots are full.

| Card | Value | Special Rule |
|------|-------|-------------|
| Pi | 3.14159... | |
| Euler's Number | 2.71828... | |
| Phi | 1.61803... (golden ratio) | |
| Tau | 6.28318... (2 x pi) | |
| The Answer | 42 | |
| 1st Perfect Number | 6 | |
| 2nd Perfect Number | 28 | |
| Symmetrical Number | 69 | |
| Negative One | -1 | |
| Zero | 0 | |
| Graham's Number | 9 (it's incomprehensibly large, so we compromise) | |
| Sheldon's Number | 73 if you already have 73; otherwise 12 | |
| Fibonacci Number | 1, with mark `"F"` | Grows each turn: 1, 2, 3, 5, 8, 13, 21, ... |
| Lucky Number | 7, plus a bonus 7 if two dice sum to 7 | |
| Cupid's Number | 29 if both dice are 3 or less; otherwise 14 | |
| Scientific Notation | 10^d where d is a dice roll | |
| Factorial | d! where d is a dice roll (1-720) | |
| Googol | 10, OR steal a power-of-10 from another player | Cannot steal from yourself |

### 3.3 Theorem Cards (7 cards)

Theorem cards have special effects that go beyond simple arithmetic.

| Card | Effect | Input |
|------|--------|-------|
| Identity Element | Does nothing | Target number |
| Closure Element | Grants **immunity** to one number for the current turn | Target number |
| Distributive Element | Copies one number to every other player's set | Any player's number |
| Commutative Element | Swaps any two numbers on the table | Any two numbers from any players |
| Pascal's Triangle | Collapses a group of adjacent (contiguous) numbers into their sum; immune numbers in the group are left untouched | A number in the group |
| Pythagorean Theorem | Replaces target with sqrt(target^2 + source^2); source is consumed | Target + source |
| Fundamental Theorem of Arithmetic | Decomposes an integer into its prime factors, placing each factor in a slot; original is consumed | Must be an integer > 1 |

---

## 4. Turn Structure

### 4.1 Simultaneous Play

Matetra uses a **simultaneous action** model. All players act during every turn — not just the defender. The flow:

```
  ┌─────────────────────────────────────────────────┐
  │                    TURN N                        │
  │                                                  │
  │  1. Defender is Player (N mod numPlayers)        │
  │                                                  │
  │  2. ALL players may:                             │
  │     - Roll dice to fill empty slots              │
  │     - Queue cards from their hand                │
  │     - Cards targeting defender's numbers are      │
  │       attacks; cards on your own numbers          │
  │       are defensive                              │
  │                                                  │
  │  3. ALL players press "end turn" when ready      │
  │                                                  │
  │  4. Once everyone is done:                       │
  │     - All queued cards resolve by precedence     │
  │     - Failed cards return to player's hand       │
  │     - Win condition is checked                   │
  │     - Hands are restocked from the deck          │
  │     - Immunity expires, Fibonacci numbers grow   │
  │     - Empty slots are auto-filled with dice      │
  │     - Turn advances                              │
  └─────────────────────────────────────────────────┘
```

### 4.2 The Queue and Precedence

When a player plays a card, it is **queued** rather than immediately applied. All queued cards from all players are collected and sorted by **precedence** (lower number resolves first). Cards with equal precedence resolve in the order they were queued.

This means:
- An immunity card (if it has lower precedence) resolves before an attack card, protecting the number
- An attack that targets a number which was consumed by an earlier card in the same queue will fail gracefully — the card is returned to the player's hand

### 4.3 The Preview System

While players are choosing their cards, the server generates a **preview state** — a virtual copy of the game where all currently queued cards have been applied. This preview is broadcast to all players so everyone can see the projected outcome of the current queue in real time.

The preview does not modify the actual game state. Only when all players end their turn does the queue apply for real.

### 4.4 Card Failure

Cards can fail at resolution time for several reasons:
- The target number became null (consumed by an earlier card in the queue)
- The target number became immune (protected by an earlier card)
- A mathematical constraint was violated (divide by zero, sqrt of negative, non-integer for prime decomposition)

When a card fails, it is **returned to the player's hand** rather than consumed. The game state is not modified by the failed card.

### 4.5 Between Turns

After all cards resolve:

1. **Scoring**: In Limit mode, any number exceeding the limit scores a point and the slot becomes empty
2. **Restocking**: Players draw cards from the deck until their hand is full (default: 6 cards). If the deck is empty, used cards are reshuffled back in
3. **Immunity expiry**: Numbers marked `"I"` become `""` (normal). Numbers marked `"FI"` become `"F"` (fibonacci only)
4. **Fibonacci growth**: Numbers marked `"F"` advance to the next Fibonacci number in the sequence
5. **Auto-fill**: Any remaining empty slots are filled with random dice rolls (1-6)

---

## 5. Input System

### 5.1 Input Codes

Each card defines an **InputsReq** string that describes what inputs it needs. Each character is one input slot:

| Code | Meaning | How It's Filled |
|------|---------|-----------------|
| `A` | Defender's player ID | Auto-filled (current turn player) |
| `U` | Card owner's player ID | Auto-filled (you) |
| `d` | Dice value (1-6) | Auto-filled (random) |
| `n` | Number slot index (0-4) | Player types it |
| `p` | Any player's ID | Player types it |

Examples:
- `AnUn` (Addition): Auto-target defender, pick their slot, auto-target self, pick your slot
- `An` (Square): Auto-target defender, pick their slot
- `pn` (Googol): Pick any player, pick their slot
- `pnpn` (Commutative): Pick player 1 + slot, pick player 2 + slot
- `A` (Summation): Auto-target defender (whole set)
- `dd` (Lucky Number): Two auto dice rolls
- *(empty)* (Pi constant): No input needed

### 5.2 Validation

Inputs are validated twice:
1. **At queue time**: When a player submits a card, inputs are checked for type correctness and bounds
2. **At resolution time**: When the card actually resolves, inputs are re-validated because the game state may have changed (e.g., a number that was valid at queue time may now be null or immune)

### 5.3 Immunity Check

After type validation, every `n` input is checked for immunity. The system looks at the number at `[player][slot]` — if its mark is `"I"` or `"FI"`, the card is rejected. This prevents any card from modifying an immune number.

---

## 6. Architecture

### 6.1 System Overview

```
┌──────────────┐     WebSocket      ┌──────────────────────┐
│  TUI Client  │◄──────────────────►│    Game Server       │
│  (Bubble Tea)│     JSON msgs      │                      │
│              │                    │  ┌──────────────┐    │
│  - Card UI   │                    │  │   Hub        │    │
│  - Number    │                    │  │  (rooms map) │    │
│    display   │                    │  └──────┬───────┘    │
│  - Input     │                    │         │            │
│    fields    │                    │  ┌──────▼───────┐    │
│              │                    │  │   Room       │    │
└──────────────┘                    │  │  - Game      │    │
                                    │  │  - Conns[]   │    │
       ...more clients...           │  └──────┬───────┘    │
                                    │         │            │
                                    │  ┌──────▼───────┐    │
                                    │  │   Engine     │    │
                                    │  │  - GameState │    │
                                    │  │  - Cards[]   │    │
                                    │  │  - Numbers[] │    │
                                    │  └──────────────┘    │
                                    └──────────────────────┘
```

### 6.2 Server Components

**Hub** — The top-level singleton. Holds all active rooms and the user authentication store. Routes connecting players to their rooms.

**Room** — One per game session. Identified by a 6-letter code. Holds the game engine instance and all player connections. Responsible for message routing and state broadcasting.

**Engine (Game)** — The game logic core. Manages the game state, validates and queues cards, processes turns, checks win conditions. All state access is mutex-protected for concurrent safety.

**Cards** — Loaded from an embedded CSV file at game start. Each card is dispatched to its specific handler function based on its method name. Card handlers operate on a `GameState` object and return an error if the operation fails.

### 6.3 Client Components

The client is a terminal UI built with **Bubble Tea** (a Go TUI framework). It maintains a local copy of the game state received from the server and renders it as an interactive display.

Key modes:
- **Normal mode**: Browse cards, view numbers, select a card to play
- **Play mode**: Fill in the required inputs for the selected card (slot indices, player IDs)
- **Dice mode**: Roll dice to fill empty slots
- **Command mode**: Special commands

### 6.4 Networking

Communication happens over **WebSocket** with JSON messages.

**Connection lifecycle:**
1. Client authenticates via `POST /auth` with username + password, receives a JWT
2. Client connects to `GET /ws?token=JWT`, receives `WELCOME` or `AUTO_REJOINED`
3. Client sends `CREATE_ROOM` or `JOIN_ROOM`
4. Game messages flow bidirectionally

**Message types (client to server):**

| Message | Purpose |
|---------|---------|
| `ADD_PLAYER` | Join the game in the room |
| `START_GAME` | Begin the game (host only) |
| `PLAY_CARD` | Queue a card with inputs |
| `ROLL_DICE` | Roll dice into empty slots |
| `FINISH_DICE` | Stop rolling |
| `END_TURN` | Signal that you're done |

**Message types (server to client):**

| Message | Purpose |
|---------|---------|
| `STATE_UPDATE` | Full game state (sent after every action) |
| `REPLY` | Success/failure response to an action |
| `ERROR` | Error message |

**Keepalive:** The server sends WebSocket pings every 30 seconds. If no pong is received within 60 seconds, the connection is considered dead. This prevents silent disconnections through proxies and load balancers.

**Reconnection:** If a player disconnects and reconnects with the same credentials, the server automatically rejoins them to their active game. Their player state (numbers, cards, position) is preserved.

### 6.5 Concurrency Model

- Each player connection runs in its own goroutine (reading messages)
- A separate goroutine per connection sends periodic pings
- All game state mutations go through a single mutex (`Game.mu`), ensuring serialized access
- All WebSocket writes per connection go through a per-connection mutex (`PlayerConnection.mu`), preventing frame corruption
- Broadcast operations hold a read lock on the connection map while sending

---

## 7. The Deck

The full deck contains 49 cards. Multiple copies of common operations ensure they appear frequently:

| Count | Cards |
|-------|-------|
| 4 copies | Addition, Subtraction, Multiplication, Division |
| 2 copies | Absolute Value, Inverse, Negative, Positive |
| 1 copy | All other cards (27 unique) |

When a player uses a card, it moves from their hand to the "used" pile (owner = -2). When the deck runs out, all used cards are reshuffled back in. Players draw up to their hand size (default: 6) at the start of each turn.

---

## 8. Immunity System

Immunity is a critical defensive mechanic. The **Closure Element** card grants immunity to one number for the duration of the current turn.

**What immunity protects against:**
- All function cards (add, multiply, sqrt, etc.) are blocked
- Aggregate operations (Sigma, Product) skip immune numbers
- Pascal's Triangle skips immune numbers when collapsing an island
- Constant cards cannot overwrite immune numbers when all slots are full
- Commutative Element cannot swap an immune number

**What immunity does NOT protect against:**
- The number's own Fibonacci growth (Fibonacci + Immune numbers still grow)
- Scoring (if the number crosses the limit, it still scores and is consumed)

**Combined marks:** A Fibonacci number that receives immunity becomes `"FI"`. When immunity expires next turn, it reverts to `"F"` and continues growing.

---

## 9. Strategy

### 9.1 As Defender
- Use binary operations (Addition, Multiplication) on your own numbers to grow them toward the limit
- Play Closure Element on your most valuable number to protect it from attackers
- The Exponential card on a number > 1 can cause explosive growth
- Fibonacci numbers are a long-term investment — protect them with immunity

### 9.2 As Attacker
- Division and Subtraction shrink the defender's numbers
- Negative flips a large positive into a large negative (far from the limit)
- Inverse turns a large number into a tiny fraction
- Square Root dramatically shrinks large numbers
- Base-10 Logarithm is devastating against large numbers (1,000,000 → 6)
- Cosine/Sine with certain dice values can multiply by near-zero

### 9.3 Advanced Plays
- **Commutative Element**: Swap your small number for an opponent's large one
- **Distributive Element**: Copy a tiny or negative number to flood everyone's sets
- **Fundamental Theorem**: Decompose a prime number and it just returns itself (1 factor = original). Decompose a composite to spread factors across your slots
- **Pascal's Triangle**: If you have 5 contiguous numbers, collapse them all into a massive sum
- **Googol**: Steal a 10^d from an opponent who played Scientific Notation
- **Summation on yourself**: Combine 5 medium numbers into one large number for scoring

---

## 10. Glossary

| Term | Definition |
|------|-----------|
| **Defender** | The player whose turn it is. Their numbers are the primary target |
| **Attacker** | Any non-defender player. They play cards against the defender's numbers |
| **Slot** | One of 5 positions in a player's number set (indexed 0-4) |
| **Null** | An empty slot (mark = `"n"`). Can be filled by dice or constants |
| **Queue** | The list of cards played this turn, waiting to resolve |
| **Precedence** | Resolution order for queued cards (lower = resolves first) |
| **Preview** | A virtual state showing what the game would look like if the queue resolved now |
| **Island** | A contiguous group of non-null numbers in a player's set (used by Pascal's Triangle) |
| **Hand** | The cards a player currently holds (default: 6) |
| **Deck** | The pool of unowned cards that players draw from |
| **Consumed** | A number or card that has been used up and removed from play |
