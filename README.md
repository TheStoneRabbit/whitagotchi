# whitagotchi

A CLI Tamagotchi with multiplayer chat. Pets are hatched and live on a central server; you care for them through an animated TUI in your terminal. Pets can DM each other, and every message gets paraphrased through the pet's personality using OpenAI — so a haiku-speaking dragon really does message you in haiku.

See [CLAUDE.md](./CLAUDE.md) for the full design spec.

## Install

If someone is hosting a whitagotchi server, the easiest install is one line:

```sh
curl -fsSL http://<their-server>:8080/install | sh
```

This drops the `whitagotchi` binary on your PATH and writes a config pointing at that server.

Then:

```sh
whitagotchi register <username>   # hatches your egg, saves your auth token
whitagotchi                       # launches the interactive TUI
```

## Using it

### TUI keys (home view)

| key | action |
|---|---|
| `f` | feed |
| `p` | play |
| `b` | bathe |
| `c` | open chat |
| `R` | reroll personality quirk |
| `r` | refresh status |
| `q` | quit |

### TUI chat

`c` → type a peer's username → `enter`. You'll see your pet and theirs side-by-side. Type and hit `enter` to send. Outgoing messages are paraphrased through your quirk before the recipient sees them. `esc` returns to the home view.

You can DM yourself (chat with your own username) to see what your quirk sounds like.

### CLI (non-interactive)

```sh
whitagotchi status                 # show your creature
whitagotchi feed | play | bathe    # care actions
whitagotchi reroll                 # roll a new personality quirk
whitagotchi chat <user>            # one-shot chat from the terminal
```

## Layout

- `client/` — Go CLI/TUI (`bubbletea` + `lipgloss`)
- `server/` — Go HTTP/WebSocket server, SQLite persistence
- `shared/` — protocol types
- `deploy/` — systemd unit + `deploy.sh` for VPS hosting
- `install.sh` — fallback installer that pulls from GitHub Releases

## Run locally for development

Requires Go 1.22+.

```sh
# server (in one terminal)
cd server && OPENAI_API_KEY=sk-... go run ./cmd/whitagotchi-server

# client (in another)
cd client && go run ./cmd/whitagotchi register myname
cd client && go run ./cmd/whitagotchi
```

Default server URL is `http://localhost:8080`. Override with `WHITAGOTCHI_SERVER`.

## Self-host the server

See [`deploy/README.md`](./deploy/README.md). The short version:

```sh
HOST=root@yourbox.com ./deploy/deploy.sh
```

This cross-compiles the server (linux/amd64) **and** client binaries (mac arm64/amd64, linux amd64), uploads everything via SSH, and restarts the systemd service. The deployed server serves a tailored install script at `/install` and the client binaries at `/bin/<file>` — so your friends can install with one curl command pointed at your box.
