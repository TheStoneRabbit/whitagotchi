package tui

import (
	"fmt"

	"github.com/whitagotchi/whitagotchi/client/internal/api"
	"github.com/whitagotchi/whitagotchi/client/internal/config"
)

// Run is a stub. The real interactive TUI (bubbletea) will land here;
// for now we just print a hint so the binary still works.
func Run(_ *api.Client, _ *config.Config) error {
	fmt.Println("TUI not implemented yet. Use:")
	fmt.Println("  whitagotchi status")
	fmt.Println("  whitagotchi feed / play / bathe")
	fmt.Println("  whitagotchi chat <user>")
	return nil
}
