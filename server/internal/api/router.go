package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/whitagotchi/whitagotchi/server/internal/chat"
	"github.com/whitagotchi/whitagotchi/server/internal/game"
	"github.com/whitagotchi/whitagotchi/server/internal/store"
	"github.com/whitagotchi/whitagotchi/shared"
)

type Server struct {
	store  *store.Store
	engine *game.Engine
	hub    *chat.Hub
}

func NewRouter(s *store.Store, e *game.Engine, h *chat.Hub) http.Handler {
	srv := &Server{store: s, engine: e, hub: h}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", srv.register)
	mux.HandleFunc("GET /status", srv.auth(srv.status))
	mux.HandleFunc("POST /action", srv.auth(srv.action))
	mux.HandleFunc("GET /chat", srv.auth(srv.chatWS))
	return mux
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req shared.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}
	if _, exists := s.store.UserByName(req.Username); exists {
		http.Error(w, "username taken", http.StatusConflict)
		return
	}
	token := genToken()
	creature := game.Hatch(req.Username)
	if err := s.store.CreateUser(&store.User{Username: req.Username, Token: token}, creature); err != nil {
		http.Error(w, "create failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, shared.RegisterResponse{
		Username: req.Username,
		Token:    token,
		Creature: *creature,
	})
}

func (s *Server) status(w http.ResponseWriter, r *http.Request, user *store.User) {
	c, ok := s.store.Creature(user.Username)
	if !ok {
		http.Error(w, "no creature", http.StatusNotFound)
		return
	}
	s.engine.AdvanceNow(c)
	s.store.UpdateCreature(c)
	writeJSON(w, shared.StatusResponse{Creature: *c})
}

func (s *Server) action(w http.ResponseWriter, r *http.Request, user *store.User) {
	var req shared.ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	c, ok := s.store.Creature(user.Username)
	if !ok {
		http.Error(w, "no creature", http.StatusNotFound)
		return
	}
	s.engine.AdvanceNow(c)
	game.ApplyAction(c, req.Action)
	s.store.UpdateCreature(c)
	writeJSON(w, shared.StatusResponse{Creature: *c})
}

// chatWS upgrade is implemented in chat_ws.go to keep this file deps-free.

type authedHandler func(http.ResponseWriter, *http.Request, *store.User)

func (s *Server) auth(h authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		user, ok := s.store.UserByToken(token)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h(w, r, user)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func genToken() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
