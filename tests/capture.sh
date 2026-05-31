#!/usr/bin/env bash
#
# Bring up the capture stack (DB + PHP/Angular app + recording proxy) and open the
# app in a browser so you can click around and record golden fixtures. Unlike
# run.sh, this leaves the stack RUNNING when it returns — tear it down yourself when
# you're done:
#
#   podman compose -f tests/docker-compose.yml down       # keeps the seeded DB
#   podman compose -f tests/docker-compose.yml down -v     # also wipes the DB
#
# Uses podman by default; override with DOCKER=docker ./tests/capture.sh

set -uo pipefail

cd "$(dirname "$0")/.."   # repo root

DOCKER="${DOCKER:-podman}"
COMPOSE="$DOCKER compose -f tests/docker-compose.yml"
URL="http://localhost:9000"

echo "==> starting stack (db + app + proxy)"
$COMPOSE up -d || exit 1

echo "==> waiting for the PHP app + seeded DB"
for i in $(seq 1 60); do
  out=$(curl -s -m5 -X POST -H 'Accept: application/json' http://localhost:8080/entity/getSchools 2>/dev/null)
  [ "${out:0:1}" = "[" ] && { echo "    ready"; break; }
  [ "$i" = 60 ] && { echo "    app never became ready" >&2; exit 1; }
  sleep 2
done

echo "==> opening $URL"
if command -v xdg-open >/dev/null 2>&1; then xdg-open "$URL" >/dev/null 2>&1 &
elif command -v open      >/dev/null 2>&1; then open "$URL" >/dev/null 2>&1 &
else echo "    (no browser opener found — visit $URL yourself)"; fi

cat <<EOF

Capture stack is running.

  * Browse the app at $URL  (the PROXY — records to tests/fixtures/)
    Do NOT use http://localhost:8080 directly; that bypasses recording.
  * Error/edge cases:   cd tests && PROXY=$URL EDGE_TERM=20261 ./edge-cases.sh
  * See recorded goldens: ls tests/fixtures/

When finished, stop the stack:
  $COMPOSE down        # keeps the seeded DB for next time
  $COMPOSE down -v     # also wipes the DB
EOF
