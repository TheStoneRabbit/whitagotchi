package main

import (
	"fmt"
	"os"

	"github.com/whitagotchi/whitagotchi/client/internal/api"
	"github.com/whitagotchi/whitagotchi/client/internal/config"
	"github.com/whitagotchi/whitagotchi/client/internal/tui"
)

const usage = `whitagotchi - cli tamagotchi

usage:
  whitagotchi register <username>     create your account + hatch an egg
  whitagotchi status                  show your creature
  whitagotchi feed | play | bathe     care actions
  whitagotchi chat <user>             open chat with another user
  whitagotchi tui                     interactive TUI (default if no args)

env:
  WHITAGOTCHI_SERVER   server URL (default http://localhost:8080)
`

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	client := api.New(cfg)

	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"tui"}
	}

	switch args[0] {
	case "register":
		if len(args) < 2 {
			fmt.Println(usage)
			os.Exit(2)
		}
		mustRun(cmdRegister(client, cfg, args[1]))
	case "status":
		mustRun(cmdStatus(client))
	case "feed", "play", "bathe":
		mustRun(cmdAction(client, args[0]))
	case "chat":
		if len(args) < 2 {
			fmt.Println(usage)
			os.Exit(2)
		}
		mustRun(cmdChat(client, args[1]))
	case "tui":
		mustRun(tui.Run(client, cfg))
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Println(usage)
		os.Exit(2)
	}
}

func mustRun(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
