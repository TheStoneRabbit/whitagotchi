package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/whitagotchi/whitagotchi/server/internal/api"
	"github.com/whitagotchi/whitagotchi/server/internal/chat"
	"github.com/whitagotchi/whitagotchi/server/internal/game"
	"github.com/whitagotchi/whitagotchi/server/internal/store"
)

func main() {
	addr := envOr("WHITAGOTCHI_ADDR", ":8080")
	dbPath := envOr("WHITAGOTCHI_DB", "whitagotchi.db")
	binDir := envOr("WHITAGOTCHI_BIN_DIR", "bin")
	publicURL := envOr("WHITAGOTCHI_PUBLIC_URL", "")

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	engine := game.NewEngine(st, game.Config{
		PaceMultiplier: envFloat("WHITAGOTCHI_PACE", 1.0),
		TickInterval:   30 * time.Second,
	})
	go engine.Run()

	hub := chat.NewHub(st, chat.Config{OpenAIKey: os.Getenv("OPENAI_API_KEY")})
	go hub.Run()

	mux := api.NewRouter(st, engine, hub, api.InstallConfig{
		BinDir:    binDir,
		PublicURL: publicURL,
	})
	log.Printf("whitagotchi-server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envFloat(k string, def float64) float64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	var f float64
	if _, err := fmt.Sscanf(v, "%f", &f); err != nil {
		return def
	}
	return f
}
