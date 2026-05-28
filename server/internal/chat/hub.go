package chat

import (
	"log"
	"sync"
	"time"

	"github.com/whitagotchi/whitagotchi/server/internal/store"
	"github.com/whitagotchi/whitagotchi/shared"
)

type Config struct {
	OpenAIKey string
}

// Hub routes 1:1 DMs between connected users.
type Hub struct {
	cfg   Config
	store *store.Store

	mu      sync.RWMutex
	clients map[string]chan shared.ChatOutbound // username -> outbound queue

	inbound chan inboundMsg
}

type inboundMsg struct {
	from string
	msg  shared.ChatInbound
}

func NewHub(s *store.Store, cfg Config) *Hub {
	return &Hub{
		cfg:     cfg,
		store:   s,
		clients: map[string]chan shared.ChatOutbound{},
		inbound: make(chan inboundMsg, 128),
	}
}

func (h *Hub) Run() {
	for m := range h.inbound {
		h.deliver(m)
	}
}

func (h *Hub) Connect(username string) <-chan shared.ChatOutbound {
	ch := make(chan shared.ChatOutbound, 16)
	h.mu.Lock()
	h.clients[username] = ch
	h.mu.Unlock()
	return ch
}

func (h *Hub) Disconnect(username string) {
	h.mu.Lock()
	if ch, ok := h.clients[username]; ok {
		close(ch)
		delete(h.clients, username)
	}
	h.mu.Unlock()
}

func (h *Hub) Send(from string, m shared.ChatInbound) {
	h.inbound <- inboundMsg{from: from, msg: m}
}

func (h *Hub) deliver(m inboundMsg) {
	sender, ok := h.store.Creature(m.from)
	if !ok {
		return
	}
	h.mu.RLock()
	dst, online := h.clients[m.msg.To]
	h.mu.RUnlock()
	if !online {
		// TODO: queue for offline delivery, or notify sender
		log.Printf("chat: %s -> %s (recipient offline)", m.from, m.msg.To)
		return
	}

	paraphrased, err := Paraphrase(h.cfg.OpenAIKey, sender.Species, sender.Quirk, m.msg.Message)
	if err != nil {
		log.Printf("paraphrase failed: %v", err)
		paraphrased = m.msg.Message
	}

	out := shared.ChatOutbound{
		From:        m.from,
		FromSpecies: sender.Species,
		FromQuirk:   sender.Quirk,
		Paraphrased: paraphrased,
		At:          time.Now().UTC().Format(time.RFC3339),
	}
	select {
	case dst <- out:
	default:
		log.Printf("chat: dropping message to %s (buffer full)", m.msg.To)
	}
}
