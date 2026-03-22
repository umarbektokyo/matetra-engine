package main

import (
	"fmt"
	"os"

	"github.com/umarbektokyo/matetra-engine/api"
)

const splash = `
                         $$\                $$\
                         $$ |               $$ |
$$$$$$\$$$$\   $$$$$$\ $$$$$$\    $$$$$$\ $$$$$$\    $$$$$$\  $$$$$$\
$$  _$$  _$$\  \____$$\\_$$  _|  $$  __$$\\_$$  _|  $$  __$$\ \____$$\
$$ / $$ / $$ | $$$$$$$ | $$ |    $$$$$$$$ | $$ |    $$ |  \__|$$$$$$$ |
$$ | $$ | $$ |$$  __$$ | $$ |$$\ $$   ____| $$ |$$\ $$ |     $$  __$$ |
$$ | $$ | $$ |\$$$$$$$ | \$$$$  |\$$$$$$$\  \$$$$  |$$ |     \$$$$$$$ |
\__| \__| \__| \_______|  \____/  \_______|  \____/ \__|      \_______|
`

func main() {
	fmt.Print(splash)
	if len(os.Args) > 1 && os.Args[1] == "start" {
		port := os.Getenv("PORT")
		if port == "" { port = "1729" }
		fmt.Println("  server v0.2")
		fmt.Printf("  POST /auth   authenticate\n")
		fmt.Printf("  GET  /ws     websocket\n")
		fmt.Printf("  port: %s\n\n", port)
		hub := api.NewHub()
		api.Start(hub, port)
	} else {
		fmt.Println("  usage: matetra-server start")
	}
}
