// Command recordproxy is an upstream-agnostic recording reverse proxy. It forwards
// every request to UPSTREAM unchanged, and for JSON API responses it tees the
// request/response pair to disk as a golden fixture for the conformance suite.
//
// It records the exact wire request the browser sends to PHP (e.g. /entity/getSchools,
// which PHP routes via .htaccess on the Accept header). The test harness maps those
// paths to the Go server's /api/... scheme at replay time.
//
// Recording rule: the response body is valid JSON AND the path is not in the
// denylist. The JSON-body check naturally excludes the HTML index page, text/calendar
// iCal, and image/png images (none parse as JSON) and, unlike a Content-Type check,
// still captures endpoints like status.php that emit JSON without the JSON header.
// The denylist drops endpoints that write or are non-deterministic (schedule/new, rmp).
//
// Config (env): UPSTREAM (required, e.g. http://app:8080), LISTEN (default :9000),
// FIXTURE_DIR (default /fixtures), DB_DUMP (default schedulemaker.gz),
// UPSTREAM_NAME (default php).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"schedulemaker/tests/internal/fixture"
)

type ctxKey int

const reqBodyKey ctxKey = 0

// denylist drops paths that are non-deterministic or mutate state even though they
// return JSON. Matched as substrings of the request path.
var denylist = []string{"/schedule/new", "rmp"}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	upstreamRaw := os.Getenv("UPSTREAM")
	if upstreamRaw == "" {
		log.Fatal("UPSTREAM is required (e.g. http://app:8080)")
	}
	upstream, err := url.Parse(upstreamRaw)
	if err != nil {
		log.Fatalf("invalid UPSTREAM %q: %v", upstreamRaw, err)
	}
	listen := env("LISTEN", ":9000")
	fixtureDir := env("FIXTURE_DIR", "/fixtures")
	dbDump := env("DB_DUMP", "schedulemaker.gz")
	upstreamName := env("UPSTREAM_NAME", "php")

	proxy := httputil.NewSingleHostReverseProxy(upstream)
	proxy.ModifyResponse = func(resp *http.Response) error {
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		// Always restore the body so the real client still gets the response.
		resp.Body = io.NopCloser(bytes.NewReader(body))

		if denied(resp.Request.URL.Path) || !json.Valid(body) {
			return nil
		}

		var reqBody []byte
		if b, ok := resp.Request.Context().Value(reqBodyKey).([]byte); ok {
			reqBody = b
		}
		f := &fixture.Fixture{
			Method:      resp.Request.Method,
			Path:        resp.Request.URL.Path,
			Query:       resp.Request.URL.RawQuery,
			Accept:      resp.Request.Header.Get("Accept"),
			ContentType: resp.Request.Header.Get("Content-Type"),
			ReqBody:     string(reqBody),
			Status:      resp.StatusCode,
			RespBody:    json.RawMessage(body),
			Meta: fixture.Meta{
				DBDump:     dbDump,
				RecordedAt: time.Now().UTC().Format(time.RFC3339),
				Upstream:   upstreamName,
			},
		}
		path, err := f.Save(fixtureDir)
		if err != nil {
			log.Printf("ERROR saving fixture for %s %s: %v", f.Method, f.Path, err)
			return nil
		}
		log.Printf("recorded %s %s -> %s", f.Method, f.Path, path)
		return nil
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Buffer the request body so we can both forward and record it.
		var body []byte
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		r = r.WithContext(context.WithValue(r.Context(), reqBodyKey, body))
		proxy.ServeHTTP(w, r)
	})

	log.Printf("recordproxy: %s -> %s, writing %s (db=%s, upstream=%s)",
		listen, upstream, fixtureDir, dbDump, upstreamName)
	if err := http.ListenAndServe(listen, handler); err != nil {
		log.Fatal(err)
	}
}

func denied(path string) bool {
	for _, d := range denylist {
		if strings.Contains(path, d) {
			return true
		}
	}
	return false
}
