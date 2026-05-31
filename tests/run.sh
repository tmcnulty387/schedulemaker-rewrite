#!/usr/bin/env bash
#
# Run the conformance suite end-to-end and tear everything down on exit / Ctrl-C.
#
#   ./tests/run.sh
#       Bring up DB + PHP app + proxy, start the Go rewrite on :8081, replay the
#       recorded goldens against Go and diff.
#
#   GO_BASE=http://localhost:9000 PATH_PREFIX="" ./tests/run.sh
#       Skip the Go server and run the sanity check against PHP itself (every
#       fixture should pass). Any GO_BASE/PATH_PREFIX is passed through to go test.
#
# Uses podman by default; set COMPOSE / DOCKER to override (e.g. DOCKER=docker).

set -uo pipefail

cd "$(dirname "$0")/.."   # repo root

DOCKER="${DOCKER:-podman}"
COMPOSE="$DOCKER compose -f tests/docker-compose.yml"
GO_PID=""

cleanup() {
  echo
  echo "==> cleaning up"
  [ -n "$GO_PID" ] && kill "$GO_PID" 2>/dev/null
  $COMPOSE down
}
trap cleanup INT TERM EXIT

echo "==> starting stack (db + app + proxy)"
$COMPOSE up -d || exit 1

echo "==> waiting for the PHP app + seeded DB"
for i in $(seq 1 60); do
  out=$(curl -s -m5 -X POST -H 'Accept: application/json' http://localhost:8080/entity/getSchools 2>/dev/null)
  [ "${out:0:1}" = "[" ] && { echo "    ready"; break; }
  [ "$i" = 60 ] && { echo "    app never became ready" >&2; exit 1; }
  sleep 2
done

# Start the Go server unless the caller pointed GO_BASE elsewhere (e.g. at PHP).
if [ -z "${GO_BASE:-}" ]; then
  echo "==> building + starting the Go server on :8081"
  ( cd rewrite && go build -o /tmp/sm-rewrite-server ./cmd/server ) || exit 1
  pushd rewrite >/dev/null
  DATABASE_SERVER=127.0.0.1 DATABASE_USER=schedulemaker DATABASE_PASS=schedulemaker \
    DATABASE_DB=schedulemaker ADDR=127.0.0.1 PORT=8081 /tmp/sm-rewrite-server &
  GO_PID=$!
  popd >/dev/null

  for i in $(seq 1 30); do
    curl -s -m2 -o /dev/null http://127.0.0.1:8081/ && { echo "    ready"; break; }
    kill -0 "$GO_PID" 2>/dev/null || { echo "    Go server exited at boot (see rewrite/internal/api/routes.go)" >&2; exit 1; }
    [ "$i" = 30 ] && { echo "    Go server never became ready" >&2; exit 1; }
    sleep 1
  done
else
  echo "==> using GO_BASE=$GO_BASE (not starting the Go server)"
fi

echo "==> running conformance suite"
( cd tests && go test -count=1 -v )
rc=$?

echo "==> done (go test exit $rc)"
exit $rc
