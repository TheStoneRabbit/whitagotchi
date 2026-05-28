package api

import (
	"encoding/json"
	"net/http"

	"github.com/whitagotchi/whitagotchi/server/internal/store"
	"github.com/whitagotchi/whitagotchi/shared"
)

// chatWS is a stub. Real WebSocket upgrade will use gorilla/websocket or nhooyr;
// adding the dep is left for the implementation pass so this scaffold builds
// without `go get`. For now, we expose an HTTP long-poll-ish endpoint to keep
// the API surface visible.
//
// Pull one queued message (blocking is not implemented yet — clients should
// retry). POST /chat with {to, message} to send.
func (s *Server) chatWS(w http.ResponseWriter, r *http.Request, user *store.User) {
	switch r.Method {
	case http.MethodGet:
		// Connect once, read one message, return. (Placeholder.)
		ch := s.hub.Connect(user.Username)
		defer s.hub.Disconnect(user.Username)
		select {
		case m, ok := <-ch:
			if !ok {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			writeJSON(w, m)
		case <-r.Context().Done():
			w.WriteHeader(http.StatusNoContent)
		}
	case http.MethodPost:
		var in shared.ChatInbound
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		s.hub.Send(user.Username, in)
		w.WriteHeader(http.StatusAccepted)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
