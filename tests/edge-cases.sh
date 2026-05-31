#!/usr/bin/env bash
#
# Mint golden fixtures for error/boundary branches the validating frontend never
# sends, by firing deliberately malformed/empty requests through the recording
# proxy (which captures whatever PHP returns as the golden).
#
# Prereqs: the compose stack is up (db + app + proxy) so the proxy is reachable
# and forwarding to PHP. See tests/README.md.
#
#   PROXY=http://localhost:9000 EDGE_TERM=20235 ./edge-cases.sh
#
# EDGE_TERM only needs to be any value; the golden records PHP's actual response
# either way, so these stay valid regardless of what's seeded. (Named EDGE_TERM to
# avoid colliding with the shell's standard TERM terminal-type variable.)

set -euo pipefail

PROXY="${PROXY:-http://localhost:9000}"
EDGE_TERM="${EDGE_TERM:-20235}"

# fire METHOD PATH [data]
fire() {
  local method="$1" path="$2" data="${3:-}"
  echo "--> $method $path ${data:+($data)}"
  if [[ "$method" == "GET" ]]; then
    curl -sS -o /dev/null -w '    %{http_code} %{content_type}\n' \
      -H 'Accept: application/json' "$PROXY$path"
  else
    curl -sS -o /dev/null -w '    %{http_code} %{content_type}\n' \
      -H 'Accept: application/json' --data "$data" "$PROXY$path"
  fi
}

# --- entity: missing required arguments (fire before any DB query) ---
fire POST /entity/getSchoolsForTerm ''
fire POST /entity/getDepartments "term=$EDGE_TERM"
fire POST /entity/getDepartments 'school=1'
fire POST /entity/getCourses "term=$EDGE_TERM"
fire POST /entity/getCourses 'department=1'
fire POST /entity/getSections ''
fire POST /entity/courseForSection ''

# --- generate: missing args and bad course format ---
fire POST /generate/getCourseOpts "term=$EDGE_TERM"
fire POST /generate/getCourseOpts 'course=CSCI-101'
fire POST /generate/getCourseOpts "course=NOT_A_COURSE&term=$EDGE_TERM"
fire POST /generate/getMatchingSchedules 'courseCount=0&nonCourseCount=0&noCourseCount=0'

# --- search: missing term, and impossible filter (no match) ---
fire POST /search/find ''
fire POST /search/find "term=$EDGE_TERM&title=zzzzznotacourse"

# --- entity lookups against ids that should not exist ---
fire POST /entity/getSections 'course=999999999'
fire POST /entity/courseForSection 'id=999999999'

echo "done — check the fixtures directory for new goldens"
