#!/usr/bin/env sh
set -eu

install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
binary="${install_dir}/tunnel"
config_dir="$HOME/.tunneld"

if [ -f "$binary" ]; then
  rm -f "$binary"
  echo "removed ${binary}"
else
  echo "tunnel binary not found at ${binary}"
fi

if [ "${REMOVE_CONFIG:-0}" = "1" ]; then
  rm -rf "$config_dir"
  echo "removed ${config_dir}"
else
  echo "kept ${config_dir}"
  echo "set REMOVE_CONFIG=1 to remove saved cli config"
fi
