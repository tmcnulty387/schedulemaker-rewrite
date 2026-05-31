# API conformance suite (PHP → Go)

Proves the Go rewrite (`rewrite/`) returns the same data as the reference PHP API
(`schedulemaker/`) for the same inputs. It works by **golden replay**: record real
PHP `request → response` pairs once, commit them, then replay each request against
the Go server and compare the JSON.

This is the executable spec for the rewrite — with the Go handlers still stubbed,
nearly every case is red; each case turns green as its handler is implemented.

## How it works

```
RECORD:  browser / edge-cases.sh ─▶ recordproxy :9000 ─▶ PHP app :8080
                                          │ tees JSON responses
                                          ▼
                                    tests/fixtures/*.json   (committed goldens)

TEST:    go test ─▶ replay request ─▶ Go server :8081 ─▶ compare vs golden
```

- **Recorded paths are the PHP wire paths** the frontend actually sends
  (`/entity/getSchools`, `/generate/getCourseOpts`, `/search/find`, `/status`,
  `/schedule/{hex}`). PHP routes those via `.htaccess` on the `Accept` header.
- The Go server serves the same endpoints under `/api/...`, so the test prepends
  `PATH_PREFIX` (default `/api`) when replaying. Set `PATH_PREFIX=""` to replay
  against PHP itself for the self-consistency sanity check.
- **Recording rule:** the proxy saves any `application/json` response except a small
  denylist. The JSON filter auto-excludes the HTML index page, `text/calendar` iCal,
  and `image/png` images.

### Scope

**Compared:** `status`, `entity/getSchools`, `entity/getSchoolsForTerm`,
`entity/getDepartments`, `entity/getCourses`, `entity/getSections`,
`entity/courseForSection`, `search/find`, `generate/getCourseOpts`,
`generate/getMatchingSchedules`, `schedule/{hex}` GET.

**Excluded:** `rmp` (external GraphQL), `img` (binary), `schedule/new` (writes + fresh
id), `schedule/{hex}/ical` (DTSTAMP/timezone non-determinism), `terms` (no PHP XHR
endpoint — injected into `index.html` server-side).

### Comparison rules (no value normalization)

Responses are compared as **parsed JSON**, so object key order and whitespace are
ignored. Everything else is exact:

- scalar **types and values** must match (`"30"` ≠ `30`, `true` ≠ `"true"`);
- all object keys must be present on both sides;
- arrays are compared as **multisets** — same elements, order-independent (so
  `generate/getMatchingSchedules` schedule ordering doesn't matter).

## Usage

### 1. Record goldens (needs PHP + proxy)

```bash
# From the repo root (podman compose or docker compose):
podman compose -f tests/docker-compose.yml up -d   # builds proxy + app on first run

# Happy path: browse the app through the proxy and click around
#   Generate / Search / Browse / Status / open a saved schedule
open http://localhost:9000

# Error/edge path:
cd tests && PROXY=http://localhost:9000 ./edge-cases.sh
```

New `tests/fixtures/*.json` files appear (re-recording the same request overwrites).
Review and commit them.

### 2. Sanity-check the harness against PHP

Every fixture should pass when replayed against the backend that produced it:

```bash
cd tests
GO_BASE=http://localhost:9000 PATH_PREFIX="" go test -v
```

### 3. Run the conformance suite against Go

```bash
# Start the rewrite against the same (published) DB:
cd rewrite && PORT=8081 DATABASE_SERVER=127.0.0.1 go run ./cmd/server &

cd ../tests && go test -v          # GO_BASE defaults to http://localhost:8081
```

Failures for unimplemented handlers are expected. Implement a handler, re-run, watch
its subtests go green.

## Layout

| Path | Purpose |
|------|---------|
| `cmd/recordproxy/` | upstream-agnostic recording reverse proxy + Dockerfile |
| `internal/fixture/` | golden file format + load/save/dedup helpers |
| `internal/compare/` | order-insensitive, type-sensitive JSON diff |
| `compat_test.go` | replay each fixture against the server, compare |
| `edge-cases.sh` | curl error/boundary requests through the proxy |
| `docker-compose.yml` | seeded DB + PHP reference app + recording proxy |
| `fixtures/` | committed goldens |

## Notes

- Goldens are valid only for the `schedulemaker.gz` DB seed (recorded in each
  fixture's `meta.dbDump`). If the seed changes, re-record. Snapshot-per-version
  tooling is intentionally out of scope for now.
- **`schedule/{hex}` GET needs `display_errors=Off`** to capture cleanly. That
  endpoint loads the AWS SDK, which emits a PHP 7.3 deprecation notice; with the dev
  image's `display_errors=On` the notice leaks into the body (`<b>Deprecated</b>…`),
  so the response isn't valid JSON and the proxy skips it. Production runs with
  `display_errors=Off` and returns clean JSON. Set it in the PHP container before
  recording this endpoint. (entity/search/generate/status don't load the SDK and are
  unaffected.)
- **The Go server must boot to replay against it.** As of writing it panics on a
  route pattern (`GET /img/schedules/{hex}.png` is invalid for Go's ServeMux — the
  wildcard can't have a `.png` suffix), so fix that before the baseline run.
- Environment overrides: `GO_BASE`, `FIXTURE_DIR`, `PATH_PREFIX` (test);
  `UPSTREAM`, `LISTEN`, `FIXTURE_DIR`, `DB_DUMP`, `UPSTREAM_NAME` (proxy).
```
