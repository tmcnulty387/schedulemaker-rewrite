// Package compat replays recorded PHP golden requests against the Go server and
// asserts the responses are equivalent. It is the executable conformance spec for
// the rewrite: as each Go handler is implemented, its cases turn green.
//
// Env:
//
//	GO_BASE      base URL of the server under test (default http://localhost:8081)
//	FIXTURE_DIR  directory of golden fixtures (default fixtures)
//	PATH_PREFIX  prefix to prepend to recorded paths (default /api). Recorded paths
//	             are the PHP wire paths (e.g. /entity/getSchools); the Go server
//	             serves them under /api. Set PATH_PREFIX="" to replay against PHP
//	             itself (the self-consistency sanity check).
//
// Run:
//
//	go test                      # vs Go at :8081
//	GO_BASE=http://localhost:9000 PATH_PREFIX="" go test   # sanity: vs PHP via proxy
package compat

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"schedulemaker/tests/internal/compare"
	"schedulemaker/tests/internal/fixture"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// targetPath maps a recorded PHP wire path to the server-under-test path.
func targetPath(prefix, path string) string {
	if prefix == "" || strings.HasPrefix(path, prefix) {
		return path
	}
	return prefix + path
}

func TestConformance(t *testing.T) {
	base := strings.TrimRight(env("GO_BASE", "http://localhost:8081"), "/")
	dir := env("FIXTURE_DIR", "fixtures")
	// PATH_PREFIX must distinguish "unset" (default /api, replaying against Go) from
	// an explicit empty string (no prefix, e.g. the sanity check against PHP itself).
	prefix := "/api"
	if v, ok := os.LookupEnv("PATH_PREFIX"); ok {
		prefix = v
	}

	fixtures, err := fixture.LoadDir(dir)
	if err != nil {
		t.Fatalf("loading fixtures from %s: %v", dir, err)
	}
	if len(fixtures) == 0 {
		t.Skipf("no fixtures in %s — record some first (see tests/README.md)", dir)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// Pre-flight: fail once with clear guidance if the target is unreachable, rather
	// than emitting a confusing connection error for every fixture.
	if resp, err := client.Get(base + "/"); err != nil {
		t.Fatalf("server under test unreachable at %s: %v\n"+
			"start the Go server (cd rewrite && PORT=8081 DATABASE_SERVER=127.0.0.1 go run ./cmd/server),\n"+
			"or for the PHP sanity check use: GO_BASE=http://localhost:9000 PATH_PREFIX=\"\" go test", base, err)
	} else {
		resp.Body.Close()
	}

	for _, nf := range fixtures {
		f := nf.Fixture
		t.Run(nf.File, func(t *testing.T) {
			u := base + targetPath(prefix, f.Path)
			if f.Query != "" {
				u += "?" + f.Query
			}

			var bodyReader io.Reader
			if f.ReqBody != "" {
				bodyReader = strings.NewReader(f.ReqBody)
			}
			req, err := http.NewRequest(f.Method, u, bodyReader)
			if err != nil {
				t.Fatalf("building request: %v", err)
			}
			accept := f.Accept
			if accept == "" {
				accept = "application/json"
			}
			req.Header.Set("Accept", accept)
			if f.ContentType != "" {
				req.Header.Set("Content-Type", f.ContentType)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("requesting %s %s: %v", f.Method, u, err)
			}
			defer resp.Body.Close()
			got, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("reading response: %v", err)
			}

			if resp.StatusCode != f.Status {
				t.Errorf("status mismatch: golden %d, actual %d\nactual body: %s",
					f.Status, resp.StatusCode, truncate(got, 500))
			}
			if err := compare.Compare(f.RespBody, got); err != nil {
				t.Errorf("%s %s\n%v", f.Method, targetPath(prefix, f.Path), err)
			}
		})
	}
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
