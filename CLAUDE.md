# Whitagotchi

A CLI Tamagotchi with multiplayer chat. Pets live on a server; users interact via a Go TUI client. Chat between pets is paraphrased through each pet's personality using OpenAI.

## Architecture

- **Client**: Go CLI binary. Animated ASCII TUI (planned: `bubbletea` + `lipgloss`). Thin client — holds only username + auth token locally at `~/.config/whitagotchi/config.toml`.
- **Server**: Go service, packaged with Docker. Authoritative for all creature state. Owns the OpenAI API key. Runs the chat hub.
- **Transport**: REST/HTTPS for actions (feed, play, status), WebSocket for chat.
- **Persistence**: server-side DB (start with SQLite, migrate later if needed).
- **Distribution**: GitHub Actions cross-compiles client to `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`. Install via `curl -fsSL …/install.sh | sh`.

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

4 stages: **egg → baby → teen → adult**. Adult form branches based on care quality (e.g. "well-cared Dragon" vs "neglected Dragon" are different adult sprites). No permadeath — poor care just locks out the good adult forms.

### Stats

- **Hunger** — restored by `feed`
- **Happiness** — restored by `play`
- **Health/Cleanliness** — restored by `bathe` / `heal`
- **Personality quirk** — not a stat that decays; a fixed trait rolled at hatch (species base + random modifier). Shapes how chat messages are paraphrased.

### Time

- Real-world clock. Stats decay continuously **on the server**, even when no client is connected.
- Default pace: ~1–2 hours egg → adult. Configurable via server config (so we can tune for testing vs production).

### Personality quirks (random modifier rolled at hatch)

Each pet gets its species' base voice **plus** one random modifier from this pool. Server stores the modifier with the pet; uses it in the OpenAI paraphrasing prompt.

1. obsessed with cheese — works cheese into any topic
2. speaks in haiku — outputs paraphrase as 5-7-5
3. constantly tired — adds yawns, trails off mid-sentence
4. dramatic — everything is the end of the world
5. pirate — "arrr", "matey", nautical metaphors
6. shakespearean — thee/thou/forsooth
7. uwu speaker — owo, replaces r/l with w
8. conspiracy theorist — questions reality in every reply
9. surfer dude — "dude", "gnarly", "totally"
10. tiny — talks about how small they are constantly
11. royalty — refers to self in third person with titles
12. food critic — rates everything 1-10
13. detective noir — gritty, smoke-stained narration
14. overly polite — apologizes, says "if it pleases you"
15. lowercase only, no punctuation — internet shitposter energy

### Multiplayer chat

- **1:1 DMs** by username: `/dm puppylover42 hello!`
- **WebSocket** transport (`gorilla/websocket` or `nhooyr/websocket`)
- **Ephemeral** — no history persisted, only delivered live to connected clients
- **Server-side paraphrasing**: server calls OpenAI with `{species base voice} + {random quirk}` to rewrite the user's message before delivering. Recipient sees the paraphrased version.

## Auth

- First run: `whitagotchi register <username>` → server returns a token, client stores it in config.
- Subsequent calls authenticate with the token.

## Repo layout (planned)

```
whitagotchi/
├── client/          # Go CLI/TUI
│   ├── cmd/
│   ├── internal/
│   └── main.go
├── server/          # Go API + chat hub
│   ├── cmd/
│   ├── internal/
│   │   ├── game/      # stats decay, evolution, hatching RNG
│   │   ├── chat/      # WS hub, paraphrasing
│   │   ├── store/     # SQLite persistence
│   │   └── api/       # HTTP handlers
│   ├── Dockerfile
│   └── main.go
├── shared/          # protocol types shared between client/server
├── install.sh       # curl|sh installer
├── .github/workflows/
│   ├── release.yml  # build binaries on tag
│   └── ci.yml
└── CLAUDE.md
```

## Open questions / future

- Whether to add trading or social features beyond DMs
- Whether to expose a `whitagotchi list-online` for discovering other users
- Rate limiting on OpenAI calls (per-user cap to control cost)
- Optional self-hosting docs (run your own server, point client at it via `--server` flag)
