#!/usr/bin/env bash
set -euo pipefail

# Builds the folio CLI and installs it on PATH, then signs in if no token is
# cached yet. Defaults to ~/.local/bin so no sudo is needed; set PREFIX to
# override (e.g. PREFIX=/usr/local/bin, which may require sudo).

CLI_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PREFIX="${PREFIX:-$HOME/.local/bin}"
BINARY="folio"

if [[ -n ${NO_COLOR:-} ]]; then
	GREEN='' YELLOW='' NC=''
else
	GREEN='\033[0;32m' YELLOW='\033[1;33m' NC='\033[0m'
fi

echo "Building $BINARY..."
(cd "$CLI_DIR" && go build -ldflags="-s -w" -o "$CLI_DIR/bin/$BINARY" ./cmd/folio)

mkdir -p "$PREFIX"
if [[ ! -w $PREFIX ]]; then
	echo -e "${YELLOW}$PREFIX is not writable. Re-run with PREFIX pointing somewhere you own, or prefix this command with sudo.${NC}" >&2
	exit 1
fi

install -m 0755 "$CLI_DIR/bin/$BINARY" "$PREFIX/$BINARY"
echo -e "${GREEN}✓${NC} installed $PREFIX/$BINARY"

if ! printf '%s' ":$PATH:" | grep -q ":$PREFIX:"; then
	echo -e "${YELLOW}$PREFIX is not on your PATH. Add this to your shell profile:${NC}"
	echo "  export PATH=\"$PREFIX:\$PATH\""
fi

CREDENTIALS="$("$PREFIX/$BINARY" config path)"
if [[ -n ${FOLIO_TOKEN:-} ]]; then
	echo "FOLIO_TOKEN is set; skipping login."
	exit 0
fi
if [[ -f $CREDENTIALS ]]; then
	echo "Already signed in. Run '$BINARY login' to switch accounts."
	exit 0
fi

# A non-interactive install (CI, a pipe) cannot answer a prompt; leave the
# binary in place and let the operator log in later.
if [[ ! -t 0 ]]; then
	echo "Not a terminal; run '$BINARY login' when you are ready."
	exit 0
fi

echo
echo "Sign in to folio:"
"$PREFIX/$BINARY" login
