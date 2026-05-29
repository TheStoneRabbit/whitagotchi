package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/whitagotchi/whitagotchi/server/internal/store"
	"github.com/whitagotchi/whitagotchi/shared"
)

// chatWS upgrades the request to a WebSocket and runs a per-user reader/writer pair.
func (s *Server) chatWS(w http.ResponseWriter, r *http.Request, user *store.User) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Printf("ws accept (%s): %v", user.Username, err)
		return
	}
	defer c.CloseNow()

	ch := s.hub.Connect(user.Username)
	defer s.hub.Disconnect(user.Username)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Writer: forward hub messages to the socket.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-ch:
				if !ok {
					return
				}
				wctx, wcancel := context.WithTimeout(ctx, 10*time.Second)
				err := wsjson.Write(wctx, c, m)
				wcancel()
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()

	// Reader: pull inbound messages off the socket.
	for {
		var in shared.ChatInbound
		if err := wsjson.Read(ctx, c, &in); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("ws read (%s): %v", user.Username, err)
			}
			c.Close(websocket.StatusNormalClosure, "")
			return
		}
		s.hub.Send(user.Username, in)
	}
}
