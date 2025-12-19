package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/umarbektokyo/matetra-engine/api"
	"github.com/umarbektokyo/matetra-engine/model"
	"github.com/umarbektokyo/matetra-engine/utils"

	"github.com/gorilla/websocket"
)

// Global state tracking variables
var CurrentGameState model.GameState
var PlayerID int = -1
var PlayerName string
var Banner string

func main() {
	log.SetFlags(0)

	cmd := os.Args
	utils.MatetraSplash()

	if len(cmd) < 2 {
		fmt.Println("usage: matetra-client <server-address>:1729")
		fmt.Println("example: matetra-client localhost:1729")
		return
	}

	serverAddr := cmd[1]

	if !strings.HasPrefix(serverAddr, "ws://") {
		serverAddr = "ws://" + serverAddr
	}

	u, err := url.Parse(serverAddr)
	if err != nil {
		log.Fatalf("error: invalid server address format: %v", err)
	}
	u.Path = "/ws"

	fmt.Printf("Attempting to connect to server at %s...\n", serverAddr)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("error: could not connect to server at %s. Is the server still running? %v", u.String(), err)
	}
	defer c.Close()
	fmt.Println("Connection successful.")

	// register the player and BLOCK until initial state is received and displayed
	if err := registerPlayer(c); err != nil {
		log.Fatalf("Registration failed: %v", err)
	}

	// start listening for server updates (now running asynchronously)
	go listenForUpdates(c)

	// Start the command interface
	commandLoop(c)
}

// ----------------------------------------------------------------------
// REGISTRATION AND INITIAL STATE SETUP
// ----------------------------------------------------------------------

func registerPlayer(c *websocket.Conn) error {
	reader := bufio.NewReader(os.Stdin)

	// get username
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// get password
	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return fmt.Errorf("username and password cannot be empty")
	}

	passwordHash := utils.Hash(password)
	fmt.Printf("Hashed password (SHA256): %s...\n", passwordHash[:8])
	PlayerName = username // Set player name globally

	// construct the ADD_PLAYER message
	payload := api.PlayerPayload{
		Name: username,
		Hash: passwordHash,
	}
	addPlayerMsg := api.Message{
		Type:    "ADD_PLAYER",
		Payload: payload,
	}

	fmt.Println("Registering player...")
	if err := c.WriteJSON(addPlayerMsg); err != nil {
		return fmt.Errorf("error sending registration request: %v", err)
	}

	// 1. Wait for server's "PLAYER_ADDED" response
	// FIX: Loop to handle unsolicited messages (like broadcasts) that might arrive before PLAYER_ADDED
	for {
		var response api.Message
		if err := c.ReadJSON(&response); err != nil {
			return fmt.Errorf("error reading registration response: %v", err)
		}

		if response.Type == "PLAYER_ADDED" {
			fmt.Printf("Success: player @%s has been added to the game!\n", username)
			break
		} else if response.Type == "ERROR" {
			errorPayloadBytes, err := json.Marshal(response.Payload)
			if err != nil {
				return fmt.Errorf("registration failed: unknown error format")
			}
			var errorData map[string]string
			json.Unmarshal(errorPayloadBytes, &errorData)
			return fmt.Errorf("registration failed: %s", errorData["message"])
		} else {
			// Ignore other messages (e.g. broadcasts for other players joining)
			// expected behavior in a simplified client
			continue
		}
	}

	// 2. CRITICAL FIX: BLOCKING WAIT FOR INITIAL STATE_UPDATE
	// The server sends the STATE_UPDATE immediately after "PLAYER_ADDED".
	var stateUpdateMsg api.Message
	if err := c.ReadJSON(&stateUpdateMsg); err != nil {
		return fmt.Errorf("error reading initial state update: %v", err)
	}

	var gameState model.GameState

	// Server sends STATE_UPDATE (old way) or PLAY_CARD_REPLY (new way, contains state)
	if stateUpdateMsg.Type == "PLAY_CARD_REPLY" {
		var reply api.CardPlayReply
		payloadBytes, _ := json.Marshal(stateUpdateMsg.Payload)
		if err := json.Unmarshal(payloadBytes, &reply); err == nil && reply.NewGameState != nil {
			gameState = *reply.NewGameState
		}
	} else if stateUpdateMsg.Type == "STATE_UPDATE" {
		statePayloadBytes, err := json.Marshal(stateUpdateMsg.Payload)
		if err != nil {
			return fmt.Errorf("error marshalling initial state payload: %v", err)
		}
		if err := json.Unmarshal(statePayloadBytes, &gameState); err != nil {
			return fmt.Errorf("error unmarshalling initial GameState: %v", err)
		}
	} else {
		return fmt.Errorf("unexpected message type after PLAYER_ADDED: %s", stateUpdateMsg.Type)
	}

	if gameState.GameID == "" {
		return fmt.Errorf("failed to retrieve valid initial game state")
	}

	CurrentGameState = gameState

	// Find PlayerID from the received state
	for i, p := range gameState.Players {
		if p.Name == username {
			PlayerID = i
			break
		}
	}
	if PlayerID == -1 {
		return fmt.Errorf("could not find registered player ID in initial state")
	}

	// Display the initial state before starting the command loop
	displayGameState(CurrentGameState)

	return nil
}

// ----------------------------------------------------------------------
// GAME STATE DISPLAY
// ----------------------------------------------------------------------

func displayGameState(gs model.GameState) {
	fmt.Print("\033[H\033[2J") // Clear terminal screen

	// Ensure PlayerID is set (should be from registerPlayer, but safe check)
	if PlayerID == -1 {
		for i, p := range gs.Players {
			if p.Name == PlayerName {
				PlayerID = i
				break
			}
		}
	}

	if len(gs.Players) == 0 || gs.Turn == -1 {
		fmt.Println("Waiting for game to start...")
		return
	}

	// Determine current player index safely
	currentPlayerIndex := gs.Turn % len(gs.Players)

	fmt.Println(Banner)
	fmt.Println("\n=====================================================================")
	fmt.Printf(" GAME: %s | TURN: %d | CURRENT PLAYER: @%s (ID: %d)\n", gs.GameID, gs.Turn, gs.Players[currentPlayerIndex].Name, currentPlayerIndex)
	fmt.Println("=====================================================================")

	// 1. Display Player Numbers
	fmt.Println("\n--- PLAYER NUMBERS ---")
	for i, p := range gs.Players {
		doneStatus := "✅ DONE"
		if i < len(gs.Done) && !gs.Done[i] {
			doneStatus = "▶️ ACTIVE"
		}

		marker := "  "
		if i == PlayerID {
			marker = ">>" // Me
		} else if i == currentPlayerIndex {
			marker = "🎯" // Turn player
		}

		numberStrings := make([]string, 5)
		if i < len(gs.Numbers) {
			for j, num := range gs.Numbers[i] {
				// Format: [Index:ValueMark]
				displayValue := num.Value.Text('g', 10)
				numberStrings[j] = fmt.Sprintf("[%d:%s%s]", j, displayValue, num.Mark)
			}
		}

		fmt.Printf("%s %s @%s (ID: %d): %s\n", marker, doneStatus, p.Name, i, strings.Join(numberStrings, " | "))
	}

	// 2. Display Player Hand
	fmt.Println("\n--- YOUR HAND ---")
	handCount := 0
	for i, card := range gs.Cards {
		if card.Owner == PlayerID {
			handCount++
			// Find the required input string from the card
			inputsReq := card.InputsReq
			fmt.Printf("  [C:%d] %s (Req: %s) -> %s\n", i, card.Name, inputsReq, card.Description)
		}
	}
	if handCount == 0 {
		fmt.Println("  (Your hand is empty)")
	}

	// 3. Display Queue
	fmt.Println("\n--- MOVE QUEUE (Pending Moves) ---")
	if len(gs.Queue) > 0 {
		queueDetails := make([]string, len(gs.Queue))
		for i, cardIndex := range gs.Queue {
			// Find the card name
			cardName := "Unknown Card"
			if cardIndex >= 0 && cardIndex < len(gs.Cards) {
				cardName = gs.Cards[cardIndex].Name
			}
			// Display the card and the inputs (if available, they should be in the state)
			var inputs string
			if cardIndex < len(gs.Cards) && len(gs.Cards[cardIndex].Inputs) > 0 {
				inputs = fmt.Sprintf("(Inputs: %v)", gs.Cards[cardIndex].Inputs)
			}
			queueDetails[i] = fmt.Sprintf("%s %s", cardName, inputs)
		}
		fmt.Printf("  %s\n", strings.Join(queueDetails, " -> "))
	} else {
		fmt.Println("  (Queue is empty)")
	}

	fmt.Println("---------------------------------------------------------------------")
	fmt.Println("\n💡 COMMANDS:")
	fmt.Println("  apply(cardIndex, [inputs...], permanent)  - Play a card (permanent=1, preview=0)")
	fmt.Println("  roll / dice                               - Roll the dice")
	fmt.Println("  turnend                                   - End your turn")
	fmt.Println("  state                                     - Refresh board")
	fmt.Println("  help                                      - Show help")
	fmt.Println("  exit                                      - Quit")
}

// ----------------------------------------------------------------------
// MESSAGE LISTENER (Async)
// ----------------------------------------------------------------------

func listenForUpdates(c *websocket.Conn) {
	for {
		var msg api.Message
		if err := c.ReadJSON(&msg); err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				fmt.Println("\nServer connection closed.")
				os.Exit(0)
			}
			log.Printf("Error reading message: %v", err)
			continue
		}

		// Handle all state updates and replies
		switch msg.Type {
		case "PLAY_CARD_REPLY":
			var reply api.CardPlayReply
			payloadBytes, _ := json.Marshal(msg.Payload)
			if err := json.Unmarshal(payloadBytes, &reply); err != nil {
				log.Printf("Error unmarshalling CardPlayReply: %v", err)
				continue
			}

			// 1. Update Global State and Redisplay
			if reply.NewGameState != nil {
				CurrentGameState = *reply.NewGameState
				displayGameState(CurrentGameState)
			}

			// 2. Display Message
			prefix := "[INFO]"
			if !reply.Success {
				prefix = "[ERROR]"
			}
			fmt.Printf("\n%s %s\n", prefix, reply.Message)
			fmt.Print(">>> ")

		case "ERROR":
			errorPayloadBytes, _ := json.Marshal(msg.Payload)
			var errorData map[string]string
			json.Unmarshal(errorPayloadBytes, &errorData)
			fmt.Printf("\n[SERVER ERROR]: %s\n", errorData["message"])
			fmt.Print(">>> ")
		case "STATE_UPDATE":
			// FIX: Handle global state updates (e.g. when other players join or turn changes)
			statePayloadBytes, err := json.Marshal(msg.Payload)
			if err != nil {
				log.Printf("Error marshalling state payload: %v", err)
				continue
			}
			var newMessageState model.GameState
			if err := json.Unmarshal(statePayloadBytes, &newMessageState); err != nil {
				log.Printf("Error unmarshalling GameState: %v", err)
				continue
			}
			CurrentGameState = newMessageState
			displayGameState(CurrentGameState)
			fmt.Print("\n>>> ")

		default:
			// Ignore unhandled types like "PLAYER_ADDED"
		}
	}
}

// ----------------------------------------------------------------------
// COMMAND INTERFACE (Blocking)
// ----------------------------------------------------------------------

func commandLoop(c *websocket.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		// Ensure the command prompt appears clearly after the state
		fmt.Printf("\n>>> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Simple parsing: separate command from arguments inside parentheses
		// e.g., apply(2, 0, 1, 1) -> ["apply", "2", "0", "1", "1"]
		parts := strings.FieldsFunc(input, func(r rune) bool {
			return r == '(' || r == ')' || r == ',' || r == ' '
		})

		if len(parts) == 0 {
			continue
		}

		command := strings.ToLower(parts[0])

		switch command {
		case "apply":
			// Usage: apply(cardIndex, input1, input2, ..., permanent)
			// Example: apply(2, 0, 1, 1)
			if len(parts) < 3 {
				fmt.Println("Usage: apply(cardIndex, input1, ..., permanent)")
				fmt.Println("Permanent: 1 for yes, 0 for preview. Example: apply(2, 0, 1, 1)")
				continue
			}

			// Parse card index
			cardIndex, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Invalid card index (must be integer).")
				continue
			}

			// Extract inputs and 'permanent' flag
			var inputs []int
			hasError := false
			for i := 2; i < len(parts)-1; i++ {
				inputVal, err := strconv.Atoi(parts[i])
				if err != nil {
					fmt.Printf("Invalid input value '%s' at position %d.\n", parts[i], i-1)
					hasError = true
					break
				}
				inputs = append(inputs, inputVal)
			}
			if hasError {
				continue // Error already reported
			}

			// Parse permanent flag
			permanentInt, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil || (permanentInt != 0 && permanentInt != 1) {
				fmt.Println("Invalid permanent flag. Use 1 for permanent, 0 for preview.")
				continue
			}

			permanent := permanentInt == 1

			sendPlayCard(c, cardIndex, inputs, permanent)

		case "turnend":
			sendTurnEnd(c)

		case "roll", "dice":
			sendDiceRoll(c)

		case "exit", "quit":
			fmt.Println("Exiting client.")
			return

		case "state":
			if CurrentGameState.GameID != "" {
				displayGameState(CurrentGameState)
			} else {
				fmt.Println("Waiting for initial game state...")
			}
		case "help":
			fmt.Println("\nAvailable Commands:")
			fmt.Println("  apply(C, I1..., P) : Play card")
			fmt.Println("  dice               : Roll dice")
			fmt.Println("  turnend            : End turn")
			fmt.Println("  state              : Refresh")
			fmt.Println("  exit               : Quit")

		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
		}
	}
}

// ----------------------------------------------------------------------
// COMMAND SENDERS
// ----------------------------------------------------------------------

func sendPlayCard(c *websocket.Conn, cardIndex int, inputs []int, permanent bool) {
	if PlayerID == -1 {
		fmt.Println("[ERROR] Player ID not yet established. Cannot move.")
		return
	}

	playCardMsg := api.Message{
		Type: "PLAY_CARD",
		Payload: api.CardPlayPayload{
			CardIndex: cardIndex,
			Inputs:    inputs,
			Permanent: permanent,
		},
	}

	if err := c.WriteJSON(playCardMsg); err != nil {
		log.Printf("Error sending PLAY_CARD: %v", err)
	}
	if !permanent {
		fmt.Println("Sent preview request. Waiting for reply...")
	} else {
		fmt.Println("Sent permanent move. Waiting for board update...")
	}
}

func sendTurnEnd(c *websocket.Conn) {
	if PlayerID == -1 {
		fmt.Println("[ERROR] Player ID not yet established. Cannot end turn.")
		return
	}

	// Note: We assume the server's API has been updated to handle PROCESS_NEXT_TURN
	turnEndMsg := api.Message{
		Type:    "PROCESS_NEXT_TURN",
		Payload: nil,
	}

	if err := c.WriteJSON(turnEndMsg); err != nil {
		log.Printf("Error sending PROCESS_NEXT_TURN: %v", err)
	}
	fmt.Println("Sent turn end request. Waiting for update...")
}

func sendDiceRoll(c *websocket.Conn) {
	if PlayerID == -1 {
		fmt.Println("[ERROR] Player ID not established.")
		return
	}

	msg := api.Message{
		Type:    "ROLL_DICE",
		Payload: nil,
	}

	if err := c.WriteJSON(msg); err != nil {
		log.Printf("Error sending ROLL_DICE: %v", err)
	}
}
