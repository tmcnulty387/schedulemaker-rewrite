// Package compare implements order-insensitive, type-sensitive deep equality for
// JSON values, used to check a Go API response against a recorded PHP golden.
//
// Comparison is on parsed JSON, so object key order and insignificant whitespace
// never matter. Everything else is exact: scalar types and values must match
// (JSON string "30" != number 30, true != "true"), all keys must be present on
// both sides, and arrays are compared as multisets (same elements, any order).
package compare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Parse decodes JSON using json.Number so numeric values keep their exact textual
// form (avoiding float64 precision loss) and stay type-distinct from JSON strings.
func Parse(raw []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

// Compare parses both bodies and returns nil if they are equivalent, or an error
// describing the first mismatch with a JSON path (e.g. "$.courses[0].maxenroll").
func Compare(want, got []byte) error {
	w, err := Parse(want)
	if err != nil {
		return fmt.Errorf("golden response is not valid JSON: %w", err)
	}
	g, err := Parse(got)
	if err != nil {
		return fmt.Errorf("actual response is not valid JSON: %w", err)
	}
	return diff("$", w, g)
}

func diff(path string, want, got any) error {
	switch w := want.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			return typeMismatch(path, want, got)
		}
		for k := range w {
			if _, present := g[k]; !present {
				return fmt.Errorf("%s: missing key %q in actual response", path, k)
			}
		}
		for k := range g {
			if _, present := w[k]; !present {
				return fmt.Errorf("%s: unexpected key %q in actual response", path, k)
			}
		}
		// Stable iteration so failures are deterministic.
		keys := make([]string, 0, len(w))
		for k := range w {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if err := diff(childPath(path, k), w[k], g[k]); err != nil {
				return err
			}
		}
		return nil

	case []any:
		g, ok := got.([]any)
		if !ok {
			return typeMismatch(path, want, got)
		}
		return diffArray(path, w, g)

	default:
		if !deepEqual(want, got) {
			return fmt.Errorf("%s: value mismatch\n  golden: %s\n  actual: %s",
				path, describe(want), describe(got))
		}
		return nil
	}
}

// diffArray matches two arrays as multisets: each golden element must pair with a
// distinct, deeply-equal actual element. Order does not matter.
func diffArray(path string, want, got []any) error {
	if len(want) != len(got) {
		return fmt.Errorf("%s: array length mismatch (golden %d, actual %d)", path, len(want), len(got))
	}
	used := make([]bool, len(got))
	for i, wv := range want {
		matched := false
		for j, gv := range got {
			if !used[j] && deepEqual(wv, gv) {
				used[j] = true
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%s: no actual element matches golden element %s%s",
				path, indexPath("", i), "\n  golden element: "+describe(wv))
		}
	}
	return nil
}

// deepEqual is the order-insensitive, type-sensitive equality used for matching
// array elements (no path tracking).
func deepEqual(a, b any) bool {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			bvv, present := bv[k]
			if !present || !deepEqual(v, bvv) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		used := make([]bool, len(bv))
		for _, x := range av {
			found := false
			for j, y := range bv {
				if !used[j] && deepEqual(x, y) {
					used[j] = true
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	case json.Number:
		bv, ok := b.(json.Number)
		return ok && av.String() == bv.String()
	default:
		return a == b
	}
}

func typeMismatch(path string, want, got any) error {
	return fmt.Errorf("%s: type mismatch\n  golden: %s\n  actual: %s",
		path, describe(want), describe(got))
}

func describe(v any) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case string:
		return fmt.Sprintf("string %q", t)
	case bool:
		return fmt.Sprintf("bool %v", t)
	case json.Number:
		return fmt.Sprintf("number %s", t.String())
	case map[string]any:
		return fmt.Sprintf("object with %d keys", len(t))
	case []any:
		return fmt.Sprintf("array of %d", len(t))
	default:
		return fmt.Sprintf("%T %v", v, v)
	}
}

func childPath(path, key string) string {
	if needsQuote(key) {
		return fmt.Sprintf("%s[%q]", path, key)
	}
	return path + "." + key
}

func indexPath(path string, i int) string { return fmt.Sprintf("%s[%d]", path, i) }

func needsQuote(key string) bool {
	if key == "" {
		return true
	}
	return strings.IndexFunc(key, func(r rune) bool {
		return !(r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	}) >= 0
}
