#!/bin/bash

set -e

# Combine .envrc files from providers, enterprise-providers, and encryption-bins
server_versions=""
provider_registries=""

shopt -s failglob
for file in /boeing-providers/.envrc.*; do
  eval "$(grep '^export ' "$file" | sed 's/^export //')"

  if [[ -n "$BOEING_SERVER_PROVIDER_REGISTRIES" ]]; then
    provider_registries+="$BOEING_SERVER_PROVIDER_REGISTRIES,"
  fi

  if [[ -n "$BOEING_SERVER_VERSIONS" ]]; then
    server_versions+="$BOEING_SERVER_VERSIONS,"
  fi
done

cat <<EOF >/boeing-providers/.envrc.providers
export BOEING_SERVER_PROVIDER_REGISTRIES="${provider_registries%,}"
export BOEING_SERVER_VERSIONS="${server_versions%,}"
EOF

rm -f /boeing-providers/.envrc.providers.*
