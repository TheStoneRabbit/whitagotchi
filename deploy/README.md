# Deploying whitagotchi-server

Target: Linux VPS, systemd, binary on port 8080 (no reverse proxy yet).

## One-time setup on the VPS

SSH into the box and run (as root or with sudo):

```sh
# 1. dedicated user + data dir
useradd --system --home /var/lib/whitagotchi --shell /usr/sbin/nologin whitagotchi
install -d -o whitagotchi -g whitagotchi /var/lib/whitagotchi

# 2. systemd unit (copy the file from this repo)
#    Put the OpenAI key in /etc/whitagotchi.env first:
echo 'OPENAI_API_KEY=sk-...' > /etc/whitagotchi.env
chmod 600 /etc/whitagotchi.env

# Then install the unit (from your laptop):
#   scp deploy/whitagotchi-server.service you@box:/tmp/
# On the box:
install -m 0644 /tmp/whitagotchi-server.service /etc/systemd/system/whitagotchi-server.service
systemctl daemon-reload
systemctl enable whitagotchi-server

# 3. open the port (ufw shown; use whatever firewall you have)
ufw allow 8080/tcp
```

## Each deploy from your laptop

```sh
HOST=you@yourbox.com ./deploy/deploy.sh
```

That script cross-compiles the linux/amd64 binary, scp's it up, installs it to `/usr/local/bin/whitagotchi-server`, and restarts the systemd service.

First deploy: the service won't start yet because the binary isn't there. Run `deploy.sh` first, then `systemctl start whitagotchi-server`.

## Letting others install the client from your server

Each deploy uploads `whitagotchi-darwin-{amd64,arm64}` and `whitagotchi-linux-amd64` to `/var/lib/whitagotchi/bin/`. The server exposes them at `/bin/<file>` and serves a tailored install script at `/install`:

```sh
curl -fsSL http://masonlapine.com:8080/install | sh
```

The script detects OS/arch, downloads the right binary to `~/.local/bin` (or `/usr/local/bin` if writable), and writes a default `config.json` pointing back at your server. After it runs:

```sh
whitagotchi register <username>
whitagotchi
```

To skip the client cross-compile during a fast server-only deploy:

```sh
SKIP_CLIENTS=1 HOST=root@yourbox ./deploy/deploy.sh
```

## Point the client at the server

On your laptop:

```sh
export WHITAGOTCHI_SERVER=http://yourbox.com:8080
whitagotchi register myname
```

Or save it in `~/Library/Application Support/whitagotchi/config.json` under `"server"`.

## Verify

```sh
# from the VPS
journalctl -u whitagotchi-server -f

# from anywhere
curl -i http://yourbox.com:8080/status
# expect 401 unauthorized — confirms it's reachable
```

## Things to know

- **Raw port 8080 means no TLS.** WebSockets work fine over `ws://`, but chat messages and your auth token go over the wire in plaintext. Fine for testing; add Caddy or nginx in front before sharing the URL with anyone.
- **State is in-memory** — restarts wipe all pets. Switching to SQLite is on the TODO list; until then a `systemctl restart` is a full reset.
- **Firewall**: if you're on a cloud provider (DigitalOcean, Hetzner, etc.), also open port 8080 in their dashboard's network/firewall settings, not just `ufw`.
