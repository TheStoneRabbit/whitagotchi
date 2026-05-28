package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/whitagotchi/whitagotchi/client/internal/api"
	"github.com/whitagotchi/whitagotchi/client/internal/config"
	"github.com/whitagotchi/whitagotchi/client/internal/render"
)

func cmdRegister(c *api.Client, cfg *config.Config, username string) error {
	resp, err := c.Register(username)
	if err != nil {
		return err
	}
	cfg.Username = resp.Username
	cfg.Token = resp.Token
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("Welcome, %s! Your egg is incubating...\n\n", resp.Username)
	fmt.Println(render.Creature(&resp.Creature))
	return nil
}

func cmdStatus(c *api.Client) error {
	resp, err := c.Status()
	if err != nil {
		return err
	}
	fmt.Println(render.Creature(&resp.Creature))
	return nil
}

func cmdAction(c *api.Client, action string) error {
	resp, err := c.Action(action)
	if err != nil {
		return err
	}
	fmt.Printf("%s'd!\n\n", action)
	fmt.Println(render.Creature(&resp.Creature))
	return nil
}

func cmdChat(c *api.Client, peer string) error {
	fmt.Printf("Chatting with %s. Type a message and hit enter. Ctrl-C to quit.\n", peer)
	// Spawn a goroutine to poll for inbound messages.
	go func() {
		for {
			m, err := c.ChatPoll()
			if err != nil {
				continue
			}
			if m == nil {
				continue
			}
			fmt.Printf("\n[%s the %s] %s\n> ", m.From, m.FromSpecies, m.Paraphrased)
		}
	}()
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if err := c.ChatSend(peer, line); err != nil {
			fmt.Fprintf(os.Stderr, "send failed: %v\n", err)
		}
	}
}
