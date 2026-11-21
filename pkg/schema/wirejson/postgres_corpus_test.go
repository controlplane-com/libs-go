package wirejson

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Transcode's whole premise is an assumption about what Postgres puts in a jsonb column's
// text output. Rather than assume, testdata/postgres_jsonb.txt IS that output: each line
// is `SELECT '<value>'::jsonb::text` run against the real server, for values chosen to hit
// everything the transcoder has to reason about — key order, exponents, control
// characters, non-ASCII, the line separators, an empty key, the float64 extremes.
//
// Regenerate it (and re-read it — the answers are the specification) with:
//
//	psql -h localhost -p 5433 -U postgres -d ds_diff_go -f /tmp/corpus.sql
//
// What it pins, in the server's own words:
//   - keys come back ordered by LENGTH then bytes, not alphabetically
//   - 1e21 is 22 digits and 1e-7 is 0.0000001, so the bytes must be re-rendered
//   - non-ASCII is raw UTF-8 and never \uXXXX, so escapes can be passed through
//   - U+2028/U+2029 are raw, and the wire needs them escaped
func TestTranscodeAgainstRealPostgresOutput(t *testing.T) {
	b, err := os.ReadFile("testdata/postgres_jsonb.txt")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(lines) < 10 {
		t.Fatalf("corpus looks truncated: %d lines", len(lines))
	}
	for _, src := range lines {
		got, err := Transcode(nil, []byte(src))
		if err != nil {
			t.Errorf("Transcode(%s): %v", src, err)
			continue
		}
		// The VALUE must be what the decode+encode pair would have produced.
		var fromTranscode, fromPair any
		if err := json.Unmarshal(got, &fromTranscode); err != nil {
			t.Errorf("Transcode produced invalid JSON for %s: %s (%v)", src, got, err)
			continue
		}
		var decoded any
		if err := Decode([]byte(src), &decoded); err != nil {
			t.Errorf("Decode(%s): %v", src, err)
			continue
		}
		pair, err := Marshal(decoded)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(pair, &fromPair); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(fromTranscode, fromPair) {
			t.Errorf("value differs\n  src:      %s\n  transcode: %s\n  pair:      %s", src, got, pair)
		}
		// And the KEY ORDER must be the input's, which is the one intentional
		// difference from the pair (whose map sorts them).
		if a, b := topKeys(src), topKeys(string(got)); !reflect.DeepEqual(a, b) {
			t.Errorf("key order not preserved\n  src: %v\n  got: %v", a, b)
		}
		// Nothing insignificant survives: the output is what the encoder emits.
		if strings.Contains(string(got), ": ") || strings.Contains(string(got), ", ") {
			// only inside strings is that legal; a crude check is enough here because
			// the corpus's strings contain neither sequence
			t.Errorf("whitespace not stripped: %s", got)
		}
	}
}

// topKeys lists an object's keys in the order they appear in the text.
func topKeys(s string) []string {
	var out []string
	depth, inStr, esc := 0, false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			switch {
			case esc:
				esc = false
			case c == '\\':
				esc = true
			case c == '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			if depth == 1 {
				// a key at the top level of the object: read it
				j := i + 1
				for j < len(s) {
					if s[j] == '\\' {
						j += 2
						continue
					}
					if s[j] == '"' {
						break
					}
					j++
				}
				// only count it if a ':' follows
				k := j + 1
				for k < len(s) && (s[k] == ' ') {
					k++
				}
				if k < len(s) && s[k] == ':' {
					out = append(out, s[i+1:j])
					i = j
					continue
				}
			}
			inStr = true
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		}
	}
	return out
}
