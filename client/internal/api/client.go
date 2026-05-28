package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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

func (c *Client) Action(action string) (*shared.StatusResponse, error) {
	body, _ := json.Marshal(shared.ActionRequest{Action: action})
	var out shared.StatusResponse
	if err := c.do("POST", "/action", body, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ChatSend(to, msg string) error {
	body, _ := json.Marshal(shared.ChatInbound{To: to, Message: msg})
	return c.do("POST", "/chat", body, nil, true)
}

func (c *Client) ChatPoll() (*shared.ChatOutbound, error) {
	var out shared.ChatOutbound
	err := c.do("GET", "/chat", nil, &out, true)
	if err != nil {
		return nil, err
	}
	if out.From == "" {
		return nil, nil
	}
	return &out, nil
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
