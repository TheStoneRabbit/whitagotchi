# whitagotchi

A CLI Tamagotchi with multiplayer chat. Pets are hatched and live on a central server; you care for them through a small Go TUI. Pets can DM each other, and every message gets paraphrased through the pet's personality via OpenAI.

See [CLAUDE.md](./CLAUDE.md) for the full design spec.

## Layout

- `client/` — Go CLI/TUI binary
- `server/` — Go API + chat hub (Docker)
- `shared/` — protocol types shared across client and server
- `install.sh` — `curl | sh` installer

## Development

Requires Go 1.22+.

```sh
# server
cd server && go run ./cmd/whitagotchi-server

# client (in another terminal)
cd client && go run ./cmd/whitagotchi
```
