package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/whitagotchi/whitagotchi/client/internal/config"
	"github.com/whitagotchi/whitagotchi/shared"
)

type Client struct {
	cfg  *config.Config
	http *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) Register(username string) (*shared.RegisterResponse, error) {
	body, _ := json.Marshal(shared.RegisterRequest{Username: username})
	var out shared.RegisterResponse
	if err := c.do("POST", "/register", body, &out, false); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Status() (*shared.StatusResponse, error) {
	var out shared.StatusResponse
	if err := c.do("GET", "/status", nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Reroll() (*shared.StatusResponse, error) {
	var out shared.StatusResponse
	if err := c.do("POST", "/reroll", nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Peer(name string) (*shared.PeerInfoResponse, error) {
	var out shared.PeerInfoResponse
	if err := c.do("GET", "/peer?name="+name, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Action(action string) (*shared.StatusResponse, error) {
	body, _ := json.Marshal(shared.ActionRequest{Action: action})
	var out shared.StatusResponse
	if err := c.do("POST", "/action", body, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// ChatConn is a thin wrapper around the WS connection with typed Send/Recv.
type ChatConn struct {
	conn *websocket.Conn
	ctx  context.Context
}

func (cc *ChatConn) Send(to, msg string) error {
	ctx, cancel := context.WithTimeout(cc.ctx, 10*time.Second)
	defer cancel()
	return wsjson.Write(ctx, cc.conn, shared.ChatInbound{To: to, Message: msg})
}

func (cc *ChatConn) Recv() (*shared.ChatOutbound, error) {
	var m shared.ChatOutbound
	if err := wsjson.Read(cc.ctx, cc.conn, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (cc *ChatConn) Close() error {
	return cc.conn.Close(websocket.StatusNormalClosure, "")
}

// Chat opens a WebSocket connection to the server. The returned ChatConn must be Close()'d.
func (c *Client) Chat(ctx context.Context) (*ChatConn, error) {
	if c.cfg.Token == "" {
		return nil, fmt.Errorf("not registered (run `whitagotchi register <username>`)")
	}
	url := httpToWS(c.cfg.Server) + "/chat?token=" + c.cfg.Token
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	return &ChatConn{conn: conn, ctx: ctx}, nil
}

func httpToWS(u string) string {
	switch {
	case strings.HasPrefix(u, "https://"):
		return "wss://" + strings.TrimPrefix(u, "https://")
	case strings.HasPrefix(u, "http://"):
		return "ws://" + strings.TrimPrefix(u, "http://")
	}
	return u
}

func (c *Client) do(method, path string, body []byte, out any, authed bool) error {
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, c.cfg.Server+path, r)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authed {
		if c.cfg.Token == "" {
			return fmt.Errorf("not registered (run `whitagotchi register <username>`)")
		}
		req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s: %s: %s", method, path, resp.Status, string(b))
	}
	if resp.StatusCode == 204 || out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
