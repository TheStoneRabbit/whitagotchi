# Whitagotchi

A CLI Tamagotchi with multiplayer chat. Pets live on a server; users interact via a Go TUI client. Chat between pets is paraphrased through each pet's personality using OpenAI.

## Architecture

- **Client**: Go CLI binary with an animated TUI (`bubbletea` + `lipgloss`). Thin client — holds only username + auth token locally at `~/.config/whitagotchi/config.json` (or `~/Library/Application Support/whitagotchi/config.json` on macOS).
- **Server**: Go service. Authoritative for all creature state. Owns the OpenAI API key. Runs the chat hub. Persists to SQLite (`modernc.org/sqlite`, pure-Go, no CGO).
- **Transport**: REST for actions (feed, play, bathe, reroll, status, peer); WebSocket for chat (`coder/websocket`).
- **Distribution**: `deploy/deploy.sh` cross-compiles the server (linux/amd64) and client binaries for darwin-{amd64,arm64} + linux-amd64, uploads them to the VPS, and restarts systemd. The server itself serves a tailored install script at `/install` and the client binaries at `/bin/<file>`. There is also a GitHub Releases workflow scaffolded in `.github/workflows/release.yml`.

## HTTP API

| method | path | auth | purpose |
|---|---|---|---|
| POST | `/register` | — | create user, hatch egg, return token + creature |
| GET | `/status` | bearer | fetch + tick the caller's creature |
| POST | `/action` | bearer | apply `feed` / `play` / `bathe` |
| POST | `/reroll` | bearer | roll a new (different) personality quirk |
| GET | `/peer?name=<u>` | bearer | public view of another user's creature (species/stage/quirk) |
| GET | `/chat` | bearer (?token=) | WebSocket upgrade for bidirectional chat |
| GET | `/install` | — | generated POSIX install script for the client |
| GET | `/bin/<file>` | — | static-served client binaries |

## Game design

### Creatures (6 total, 3 rarity tiers)

| Creature   | Tier      | Hatch %  | Theme     |
|------------|-----------|----------|-----------|
| Puppy      | Common    | 25%      | Cute      |
| Kitten     | Common    | 25%      | Cute      |
| Glitch-bot | Rare      | 17.5%    | Robot     |
| Slimeling  | Rare      | 17.5%    | Weird     |
| Dragon     | Legendary | 7.5%     | Mythical  |
| Phoenix    | Legendary | 7.5%     | Mythical  |

### Evolution

4 stages: **egg → baby → teen → adult**. Adult form branches based on care quality (thriving / neutral / neglect). No permadeath — poor care just locks out the good adult forms.

### Stats

- **Hunger** — restored by `feed`
- **Happiness** — restored by `play`
- **Cleanliness** — restored by `bathe`
- **Personality quirk** — a fixed trait rolled at hatch; can be rerolled. Shapes how chat messages are paraphrased.

### Time

- Real-world clock. Stats decay continuously **on the server**, even when no client is connected.
- Default pace: ~1–2 hours egg → adult. Configurable via `WHITAGOTCHI_PACE` (1.0 = default; 10.0 = 10× faster for testing).

### Personality quirks (random, rolled at hatch or via `/reroll`)

15 modifiers in the pool, defined in `shared/creature.go` (`AllQuirks`) with descriptions in `server/internal/chat/paraphrase.go` (`quirkDescription`):

cheese · haiku · tired · dramatic · pirate · shakespeare · uwu · conspiracy · surfer · tiny · royalty · foodcritic · noir · polite · lowercase

The reroll endpoint guarantees the new quirk is different from the current one.

### Multiplayer chat

- **1:1 DMs** by username over WebSocket.
- **Ephemeral** — no history persisted; messages only delivered to currently-connected recipients. If the peer is offline the message is dropped (logged server-side).
- **Server-side paraphrasing**: server calls OpenAI (default `gpt-4o-mini`) with a system prompt built from the sender's species + quirk. Recipient sees the paraphrased version; sender sees their own original text in the local chat log.
- **Fallback**: if `OPENAI_API_KEY` is unset or the call fails, the original message is delivered unchanged and a `paraphrase failed:` line is logged.

## Auth

- `whitagotchi register <username>` → server returns a token, client persists it.
- All authed endpoints accept either `Authorization: Bearer <token>` or `?token=...` (for WS).

## Server env vars

| var | default | purpose |
|---|---|---|
| `WHITAGOTCHI_ADDR` | `:8080` | listen address |
| `WHITAGOTCHI_DB` | `whitagotchi.db` | SQLite path |
| `WHITAGOTCHI_PACE` | `1.0` | scale factor for stat decay + stage progression |
| `WHITAGOTCHI_BIN_DIR` | `bin` | directory of client binaries served at `/bin/` |
| `WHITAGOTCHI_PUBLIC_URL` | (derived from request) | base URL baked into the `/install` script (set when behind TLS/reverse proxy) |
| `OPENAI_API_KEY` | — | required for chat paraphrasing |

## Repo layout

```
whitagotchi/
├── client/                          # Go CLI/TUI
│   ├── cmd/whitagotchi/
│   └── internal/{api,config,render,tui}/
├── server/                          # Go API + chat hub
│   ├── cmd/whitagotchi-server/
│   ├── internal/
│   │   ├── api/                       # HTTP router + WS upgrade + /install
│   │   ├── chat/                      # WS hub, OpenAI paraphrasing
│   │   ├── game/                      # stats decay, evolution, hatch RNG, reroll
│   │   └── store/                     # SQLite persistence
│   └── Dockerfile
├── shared/                          # protocol types shared by client/server
├── deploy/                          # systemd unit, deploy.sh, deploy README
├── install.sh                       # fallback curl|sh installer (GitHub Releases)
└── .github/workflows/{ci,release}.yml
```

## Deployment (current)

- Hosted on a Linux VPS as a systemd service. See `deploy/README.md` for one-time setup.
- `deploy/deploy.sh` builds + ships server + client binaries; `SKIP_CLIENTS=1` for fast server-only iteration.
- SQLite DB lives at `/var/lib/whitagotchi/whitagotchi.db` and survives restarts.
- Currently exposed on raw `http://`/`ws://` port 8080 — no TLS yet. Tokens and chat are plaintext on the wire; fine for testing, not for sharing wider.

## Open / future

- nginx + Let's Encrypt in front so `wss://` is real
- Offline DM queue so missed messages still deliver
- Rate limiting on `/reroll` and OpenAI calls (cost control)
- Discovery (`whitagotchi list-online`) or friend list
- More expressive ASCII art (multiple frames per stage instead of programmatic blink)
- Stronger quirk prompts / model upgrade if paraphrasing feels too mild
