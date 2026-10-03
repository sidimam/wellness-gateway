#!/bin/sh
# Parte come root solo per sistemare i permessi di /config, poi scende a nobody:users (99:100, convenzione Unraid).
dir="${CONFIG_DIR:-/config}"
mkdir -p "$dir" 2>/dev/null
chown -R 99:100 "$dir" 2>/dev/null || true
exec su-exec 99:100 /usr/local/bin/wellness-gateway "$@"
