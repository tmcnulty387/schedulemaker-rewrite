// Package fixture defines the on-disk format for recorded PHP request/response
// pairs ("goldens") and helpers shared by the recording proxy and the test suite.
package fixture

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Meta records the provenance of a golden so we know what it is valid against.
type Meta struct {
	DBDump     string `json:"dbDump"`     // DB seed the response was captured against
	RecordedAt string `json:"recordedAt"` // RFC3339 capture time
	Upstream   string `json:"upstream"`   // which backend produced it (e.g. "php")
}

// Fixture is one recorded request and the response the reference backend gave.
// RespBody holds the raw JSON bytes the backend returned; comparison parses it,
// so object key order and whitespace are irrelevant.
type Fixture struct {
	Method      string          `json:"method"`
	Path        string          `json:"path"`
	Query       string          `json:"query"`
	ContentType string          `json:"contentType,omitempty"`
	ReqBody     string          `json:"reqBody,omitempty"`
	Status      int             `json:"status"`
	RespBody    json.RawMessage `json:"respBody"`
	Meta        Meta            `json:"meta"`
}

// isForm reports whether the request body is URL-encoded form data, which we can
// canonicalize (sort) for stable dedup keys.
func (f *Fixture) isForm() bool {
	return strings.HasPrefix(f.ContentType, "application/x-www-form-urlencoded")
}

// canonical returns query and body strings with parameters sorted by key, so the
// same logical request dedups regardless of the order params happened to arrive in.
func canonicalParams(raw string) string {
	v, err := url.ParseQuery(raw)
	if err != nil {
		return raw
	}
	return v.Encode() // url.Values.Encode sorts by key
}

// Key is the dedup identity of a request: method + path + canonical query + body.
func (f *Fixture) Key() string {
	body := f.ReqBody
	if f.isForm() {
		body = canonicalParams(f.ReqBody)
	}
	return f.Method + " " + f.Path + "?" + canonicalParams(f.Query) + "\n" + body
}

// slug turns "POST /api/entity/getCourses" into "post-api-entity-getcourses".
func slug(method, path string) string {
	s := strings.ToLower(method + path)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// FileName is a stable, human-readable filename derived from the request: a slug
// of method+path plus a short hash of the full key (so distinct params don't collide).
func (f *Fixture) FileName() string {
	sum := sha256.Sum256([]byte(f.Key()))
	return fmt.Sprintf("%s-%s.json", slug(f.Method, f.Path), hex.EncodeToString(sum[:])[:8])
}

// Save writes the fixture to dir as indented JSON under its FileName.
func (f *Fixture) Save(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, f.FileName())
	return path, os.WriteFile(path, append(data, '\n'), 0o644)
}

// Load reads a single fixture file.
func Load(path string) (*Fixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f Fixture
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &f, nil
}

// Named pairs a loaded fixture with the file it came from (for test names).
type Named struct {
	File    string
	Fixture *Fixture
}

// LoadDir loads every *.json fixture in dir, sorted by filename for stable order.
func LoadDir(dir string) ([]Named, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []Named
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		f, err := Load(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, Named{File: e.Name(), Fixture: f})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].File < out[j].File })
	return out, nil
}
