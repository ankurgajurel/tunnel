#!/usr/bin/env bash
set -euo pipefail

PORT="${PORT:-5050}"
HOST="${HOST:-127.0.0.1}"

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

cat > "$TMP_DIR/index.html" <<HTML
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <title>Tunnel Test Endpoint</title>
  </head>
  <body>
    <h1>Tunnel test endpoint is running</h1>
    <p>Served from http://${HOST}:${PORT}</p>
  </body>
</html>
HTML

echo "Starting test endpoint at http://${HOST}:${PORT}"
echo "Press Ctrl+C to stop."

cd "$TMP_DIR"
python3 -m http.server "$PORT" --bind "$HOST"
