#!/usr/bin/env bash
set -euo pipefail

# Runs the CLI against a throwaway PocketBase: boots the server, authenticates
# as the seed user, runs the command, then stops the server. Dev only; the
# credentials below are the ones `go run . seed` creates.

CLI_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE_DIR="$(cd "$CLI_DIR/../base" && pwd)"

URL="${FOLIO_URL:-http://127.0.0.1:8090}"
SEED_EMAIL="${FOLIO_DEV_EMAIL:-user@test.com}"
SEED_PASSWORD="${FOLIO_DEV_PASSWORD:-pass@test}"

if curl -fsS -m 2 "$URL/api/health" > /dev/null 2>&1; then
	echo "folio: a server is already listening on ${URL#http://}." >&2
	echo "       This script starts and stops its own; stop that one first," >&2
	echo "       or point FOLIO_URL elsewhere." >&2
	exit 1
fi

SERVER_PID=""
cleanup() {
	if [[ -n $SERVER_PID ]] && kill -0 "$SERVER_PID" 2> /dev/null; then
		kill "$SERVER_PID" 2> /dev/null || true
		wait "$SERVER_PID" 2> /dev/null || true
	fi
}
trap cleanup EXIT INT TERM

SERVER_LOG="$(mktemp -t folio-dev-server)"
(cd "$BASE_DIR" && go run . serve --http="${URL#http://}") > "$SERVER_LOG" 2>&1 &
SERVER_PID=$!

for _ in $(seq 1 100); do
	if curl -fsS -m 1 "$URL/api/health" > /dev/null 2>&1; then
		break
	fi
	if ! kill -0 "$SERVER_PID" 2> /dev/null; then
		echo "folio: the server exited before becoming healthy:" >&2
		cat "$SERVER_LOG" >&2
		rm -f "$SERVER_LOG"
		exit 1
	fi
	sleep 0.3
done

if ! curl -fsS -m 1 "$URL/api/health" > /dev/null 2>&1; then
	echo "folio: the server did not become healthy in time:" >&2
	cat "$SERVER_LOG" >&2
	rm -f "$SERVER_LOG"
	exit 1
fi
rm -f "$SERVER_LOG"

if [[ -z ${FOLIO_TOKEN:-} ]]; then
	FOLIO_TOKEN="$(
		curl -fsS -X POST "$URL/api/collections/users/auth-with-password" \
			-H 'Content-Type: application/json' \
			-d "{\"identity\":\"$SEED_EMAIL\",\"password\":\"$SEED_PASSWORD\"}" \
			| sed -n 's/.*"token":"\([^"]*\)".*/\1/p'
	)"
	if [[ -z $FOLIO_TOKEN ]]; then
		echo "folio: could not authenticate as $SEED_EMAIL." >&2
		echo "       Run 'cd apps/base && go run . seed' to create the demo user." >&2
		exit 1
	fi
	export FOLIO_TOKEN
fi

export FOLIO_URL="$URL"
cd "$CLI_DIR"
go run ./cmd/folio "$@"
