# tunnel

self-hosted http tunnels for exposing a local web app through your own vps and domain.

```txt
localhost:3000 -> https://quiet-forest.tunnel.example.com
```

## local test

```sh
TUNNEL_SERVER_TOKEN=secret go run ./cmd/tunneld
```

```sh
./scripts/test_endpoint.sh
```

```sh
TUNNEL_TOKEN=secret go run ./cmd/tunnel http 5050
```

open the printed url:

```txt
http://quiet-forest.localhost:8080
```

## dns

point both records to your vps:

```txt
A  tunnel.ankurgajurel.com.np      <VPS_IP>
A  *.tunnel.ankurgajurel.com.np    <VPS_IP>
```

## env

```env
TUNNEL_HTTP_ADDR=:8080
TUNNEL_BASE_DOMAIN=tunnel.ankurgajurel.com.np
TUNNEL_PUBLIC_URL=https://tunnel.ankurgajurel.com.np
TUNNEL_SERVER_TOKEN=change-me
```

## self-host with your own reverse proxy

run only `tunneld`:

```sh
docker run -d \
  --name tunneld \
  --restart unless-stopped \
  -e TUNNEL_HTTP_ADDR=:8080 \
  -e TUNNEL_BASE_DOMAIN=tunnel.ankurgajurel.com.np \
  -e TUNNEL_PUBLIC_URL=https://tunnel.ankurgajurel.com.np \
  -e TUNNEL_SERVER_TOKEN=change-me \
  -p 127.0.0.1:8080:8080 \
  ghcr.io/ankurgajurel/tunneld:latest
```

caddy example:

```caddy
tunnel.ankurgajurel.com.np, *.tunnel.ankurgajurel.com.np {
	reverse_proxy 127.0.0.1:8080
}
```

wildcard https with cloudflare needs caddy dns-01 support.

## self-host with docker compose

```sh
cp .env.example .env
docker compose -f deploy/docker-compose.yml up -d
```

logs:

```sh
docker compose -f deploy/docker-compose.yml logs -f
```

## connect cli

download the `tunnel` binary from github releases, or build locally:

```sh
curl -fsSL https://raw.githubusercontent.com/ankurgajurel/tunnel/master/scripts/install.sh | sh
```

uninstall:

```sh
curl -fsSL https://raw.githubusercontent.com/ankurgajurel/tunnel/master/scripts/uninstall.sh | sh
```

```sh
go build -o tunnel ./cmd/tunnel
```

login once:

```sh
tunnel login
```

check what the cli will use:

```sh
tunnel config
```

then expose a local port:

```sh
tunnel http 3000
```

or override config for one run:

```sh
tunnel http 3000 --server-url http://localhost:8080 --token secret --workers 2
```

clear saved cli config:

```sh
tunnel logout
```

or use env vars:

```sh
TUNNEL_SERVER_URL=https://tunnel.ankurgajurel.com.np \
TUNNEL_TOKEN=change-me \
go run ./cmd/tunnel http 3000
```

the cli prints:

```txt
https://quiet-forest.tunnel.ankurgajurel.com.np
```

## release binaries

push a tag to build cli binaries:

```sh
make release VERSION=v0.1.0
```

## limits

- http only
- in-memory tunnel state
- shared token auth
- no dashboard
- no tcp tunnels
- agent transport uses websocket workers
