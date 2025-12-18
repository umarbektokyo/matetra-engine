package main

import (
	"fmt"
	"log"
	"os"

	"github.com/umarbektokyo/matetra-engine/api"
	"github.com/umarbektokyo/matetra-engine/engine"
	"github.com/umarbektokyo/matetra-engine/utils"
)

func main() {
	cmd := os.Args
	if len(cmd) == 1 {
		clientSplash()
		return
	}

	switch cmd[1] {
	case "start":
		title := "Wonderful Game"
		if len(cmd) > 2 {
			title = cmd[2]
		}
		utils.MatetraSplash()
		game := engine.New(title)
		log.Printf("loading card deck...")
		game.LoadCards()
		log.Printf("deck loaded with %d cards", len(game.State.Cards))

		apiServer := api.New(game)
		apiServer.Start()
	default:
		fmt.Println(cmd[1] + " not recognised.")
		clientSplash()
	}
}

func clientSplash() {
	utils.MatetraSplash()
	fmt.Println("to start a game:")
	fmt.Println("	matetra-server start <game-title>")
	fmt.Println(" ex: matetra-server start WonderfulGame")
}
