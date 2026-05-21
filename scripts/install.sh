#!/usr/bin/env sh
set -eu

repo="${TUNNEL_REPO:-ankurgajurel/tunnel}"
version="${TUNNEL_VERSION:-latest}"
install_dir="${INSTALL_DIR:-$HOME/.local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  linux|darwin) ;;
  *)
    echo "unsupported os: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "unsupported arch: $arch" >&2
    exit 1
    ;;
esac

asset="tunnel-${os}-${arch}"
if [ "$version" = "latest" ]; then
  url="https://github.com/${repo}/releases/latest/download/${asset}"
else
  url="https://github.com/${repo}/releases/download/${version}/${asset}"
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

echo "downloading ${url}"
curl -fsSL "$url" -o "$tmp"
chmod +x "$tmp"

mkdir -p "$install_dir"
mv "$tmp" "${install_dir}/tunnel"

echo "installed tunnel to ${install_dir}/tunnel"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "add ${install_dir} to PATH to run: tunnel" ;;
esac
