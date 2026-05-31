#!/usr/bin/env bash
#
# Bring up the capture stack (DB + PHP/Angular app + recording proxy) and open the
# app in a browser so you can click around and record golden fixtures. The script
# STAYS RUNNING; press Ctrl-C to stop and tear the containers down. The DB volume is
# preserved (down without -v), so the seed isn't re-imported next time.
#
# Uses podman by default; override with DOCKER=docker ./tests/capture.sh

set -uo pipefail

cd "$(dirname "$0")/.."   # repo root

DOCKER="${DOCKER:-podman}"
COMPOSE="$DOCKER compose -f tests/docker-compose.yml"
URL="http://localhost:9000"

cleaned=0
cleanup() {
  [ "$cleaned" = 1 ] && return
  cleaned=1
  echo
  echo "==> stopping stack (keeping the DB volume)"
  $COMPOSE down   # no -v: preserve db-data so we don't re-seed next run
}
trap 'cleanup; exit 0' INT TERM
trap cleanup EXIT

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

Press Ctrl-C to stop and tear the stack down (the DB volume is preserved).
EOF

# Stay up until interrupted; the trap tears the stack down on Ctrl-C.
while true; do sleep 1; done
